package murecom

// this file implement a controller for recommending music by emotion.

import (
	"errors"
	"fmt"
	"musicstore/model"
	"net/http"

	"github.com/cdfmlr/crud/log"
	"github.com/cdfmlr/crud/orm"
	"github.com/gin-gonic/gin"
)

type MurecomRequest struct {
	model.Emotion
	Limit int
}

type MurecomResponse struct {
	Tracks []*model.Track
}

// GetMurecom handles: GET /murecom
//
// Query (Notice Capitalization):
//
//   - Valence: float64, [0, 1]
//   - Arousal: float64, [0, 1]
//   - Limit: int, [1, 100], default 3
//
// Response:
//
//   - 200: OK: {tracks: [{track1}, {track2}, ...}]}
//   - 400: Bad Request: {error: "bad request"}
//   - 422: Unprocessable Entity: {error: "unprocessable entity"}
//   - 500: Internal Server Error: {error: "internal server error"}
func GetMurecom(c *gin.Context) {
	req := new(MurecomRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateMurecomRequest(c, req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	tracks, err := murecom(req.Emotion, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tracks": tracks})
}

func validateMurecomRequest(c *gin.Context, req *MurecomRequest) error {
	_, hasValence := c.GetQuery("Valence")
	_, hasArousal := c.GetQuery("Arousal")
	if !hasValence && !hasArousal {
		return errors.New("emotion query (Valence and Arousal) are required")
	}

	if req.Valence < 0 || req.Valence > 1 {
		return errors.New("valence should be in [0, 1]")
	}
	if req.Arousal < 0 || req.Arousal > 1 {
		return errors.New("arousal should be in [0, 1]")
	}
	if req.Limit == 0 { // default
		req.Limit = 3
	} else if req.Limit < 1 || req.Limit > 100 {
		return errors.New("query Limit should be in [1, 100]")
	}
	return nil
}

// murecom is the core of the murecom API.
// It returns a list of tracks that match the emotion.
//
// The algorithm is:
//
//   - Retrieval: abs(valence - ?) < 0.3 && abs(arousal - ?) < 0.3
//   - Scoring: distance(valence, arousal) = sqrt((valence - ?)^2 + (arousal - ?)^2)
//   - Re-ranking: N/A
//   - Limit: limit
//
// It's implemented by some SQL magic.
func murecom(emotion model.Emotion, limit int) ([]*model.Track, error) {
	fmt.Println("[DBG] murecom: emotion =", emotion, ", limit =", limit)
	// build SQL
	sql := `
		SELECT * FROM tracks
		WHERE
			ABS(valence - ?) < 0.3
			AND ABS(arousal - ?) < 0.3
		ORDER BY 
			SQRT(POW(valence - ?, 2) + POW(arousal - ?, 2))
		LIMIT ?
	`
	// execute SQL
	tracks := make([]*model.Track, 0)
	err := orm.DB.Raw(sql,
		emotion.Valence, emotion.Arousal, // WHERE
		emotion.Valence, emotion.Arousal, // ORDER BY
		limit, // LIMIT
	).Scan(&tracks).Error

	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}

	return tracks, nil
}
