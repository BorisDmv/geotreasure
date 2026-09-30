package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"geotreasure/internal/auth"
	"geotreasure/internal/geo"
	"geotreasure/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type treasureRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Hint        string  `json:"hint"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
}

func validCoords(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}

// CreateTreasure buries a new treasure owned by the authenticated user.
func (h *Handler) CreateTreasure(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req treasureRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 120 {
		writeError(w, http.StatusBadRequest, "name is required and must be under 120 characters")
		return
	}
	if !validCoords(req.Lat, req.Lng) {
		writeError(w, http.StatusBadRequest, "latitude/longitude out of range")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var t models.Treasure
	err := h.DB.QueryRow(ctx,
		`INSERT INTO treasures (owner_id, name, description, hint, lat, lng)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, owner_id, name, description, hint, lat, lng, created_at`,
		userID, req.Name, req.Description, req.Hint, req.Lat, req.Lng,
	).Scan(&t.ID, &t.OwnerID, &t.Name, &t.Description, &t.Hint, &t.Lat, &t.Lng, &t.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create treasure")
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

// ListTreasures returns all treasures, optionally filtered by a text query and/or
// a "near" geo search (lat, lng, radius in meters).
func (h *Handler) ListTreasures(w http.ResponseWriter, r *http.Request) {
	// Viewer is optional: works for both logged-in and anonymous requests.
	viewerID, _ := auth.UserID(r.Context())

	q := strings.TrimSpace(r.URL.Query().Get("q"))

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Build the base query. We always join owner and aggregate finds.
	args := []any{viewerID}
	where := []string{}

	if q != "" {
		args = append(args, "%"+q+"%")
		where = append(where, "(t.name ILIKE $"+strconv.Itoa(len(args))+
			" OR u.username ILIKE $"+strconv.Itoa(len(args))+")")
	}

	// Optional bounding-box pre-filter for "near" searches.
	near, radius, hasNear := parseNear(r)
	if hasNear {
		minLat, maxLat, minLng, maxLng := geo.BoundingBox(near.lat, near.lng, radius)
		args = append(args, minLat, maxLat, minLng, maxLng)
		n := len(args)
		where = append(where,
			"t.lat BETWEEN $"+strconv.Itoa(n-3)+" AND $"+strconv.Itoa(n-2)+
				" AND t.lng BETWEEN $"+strconv.Itoa(n-1)+" AND $"+strconv.Itoa(n))
	}

	sql := `
		SELECT t.id, t.owner_id, u.username, t.name, t.description, t.hint,
		       t.lat, t.lng, t.created_at,
		       COUNT(f.user_id) AS found_count,
		       BOOL_OR(f.user_id = $1) AS found_by_me
		FROM treasures t
		JOIN users u ON u.id = t.owner_id
		LEFT JOIN finds f ON f.treasure_id = t.id`
	if len(where) > 0 {
		sql += " WHERE " + strings.Join(where, " AND ")
	}
	sql += `
		GROUP BY t.id, u.username
		ORDER BY t.created_at DESC`

	rows, err := h.DB.Query(ctx, sql, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list treasures")
		return
	}
	defer rows.Close()

	treasures := make([]models.Treasure, 0)
	for rows.Next() {
		var t models.Treasure
		var foundByMe *bool
		if err := rows.Scan(&t.ID, &t.OwnerID, &t.OwnerName, &t.Name, &t.Description,
			&t.Hint, &t.Lat, &t.Lng, &t.CreatedAt, &t.FoundByCount, &foundByMe); err != nil {
			writeError(w, http.StatusInternalServerError, "could not read treasures")
			return
		}
		t.FoundByMe = foundByMe != nil && *foundByMe

		if hasNear {
			d := geo.HaversineMeters(near.lat, near.lng, t.Lat, t.Lng)
			// Exact distance filter after the bounding-box pre-filter.
			if d > radius {
				continue
			}
			t.DistanceMeters = &d
		}
		treasures = append(treasures, t)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "could not read treasures")
		return
	}

	writeJSON(w, http.StatusOK, treasures)
}

// findRadiusMeters is how close a hunter must be to mark a treasure as found.
// Kept generous to absorb consumer GPS error.
const findRadiusMeters = 50.0

type findRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// FindTreasure marks the given treasure as found by the authenticated user,
// but only if their reported location is within findRadiusMeters of it.
func (h *Handler) FindTreasure(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid treasure id")
		return
	}

	var req findRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "your current location (lat, lng) is required")
		return
	}
	if !validCoords(req.Lat, req.Lng) {
		writeError(w, http.StatusBadRequest, "latitude/longitude out of range")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Verify the treasure exists and check how far the hunter is from it.
	var tLat, tLng float64
	err = h.DB.QueryRow(ctx,
		`SELECT lat, lng FROM treasures WHERE id = $1`, id,
	).Scan(&tLat, &tLng)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "treasure not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not look up treasure")
		return
	}

	distance := geo.HaversineMeters(req.Lat, req.Lng, tLat, tLng)
	if distance > findRadiusMeters {
		writeJSON(w, http.StatusForbidden, map[string]any{
			"error":          "you are too far from this treasure to claim it",
			"distanceMeters": distance,
			"requiredMeters": findRadiusMeters,
		})
		return
	}

	// ON CONFLICT makes "find" idempotent: finding twice is a no-op.
	tag, err := h.DB.Exec(ctx,
		`INSERT INTO finds (treasure_id, user_id)
		 VALUES ($1, $2)
		 ON CONFLICT (treasure_id, user_id) DO NOTHING`,
		id, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not mark treasure found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"found":          true,
		"alreadyHad":     tag.RowsAffected() == 0,
		"distanceMeters": distance,
	})
}

// DeleteTreasure lets the owner remove their own treasure.
func (h *Handler) DeleteTreasure(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid treasure id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tag, err := h.DB.Exec(ctx,
		`DELETE FROM treasures WHERE id = $1 AND owner_id = $2`, id, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete treasure")
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "treasure not found or not yours")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type nearPoint struct {
	lat float64
	lng float64
}

func parseNear(r *http.Request) (nearPoint, float64, bool) {
	q := r.URL.Query()
	latStr := q.Get("lat")
	lngStr := q.Get("lng")
	if latStr == "" || lngStr == "" {
		return nearPoint{}, 0, false
	}
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lng, err2 := strconv.ParseFloat(lngStr, 64)
	if err1 != nil || err2 != nil || !validCoords(lat, lng) {
		return nearPoint{}, 0, false
	}
	radius := 5000.0 // default 5 km
	if rs := q.Get("radius"); rs != "" {
		if v, err := strconv.ParseFloat(rs, 64); err == nil && v > 0 {
			radius = v
		}
	}
	return nearPoint{lat: lat, lng: lng}, radius, true
}
