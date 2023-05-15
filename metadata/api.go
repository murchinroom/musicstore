package metadata

// This file provides some APIs for other packages to use.
// Saves them from constructing HTTP requests OR directly
// talking to the crud/service or even crud/orm.

import (
	"context"
	"musicstore/model"

	"github.com/cdfmlr/crud/service"
)

// TrackExists checks if the track exists in the metadata database.
func TrackExists(ctx context.Context, track *model.Track) bool {
	cnt, err := service.Count[model.Track](ctx,
		service.FilterBy("name", track.Name),
		service.FilterBy("artist", track.Artist))

	if err != nil {
		logger.WithContext(ctx).
			WithField("name", track.Name).
			WithField("artist", track.Artist).
			WithError(err).
			Error("TrackExists: failed to select tracks")
		return false
	}

	return cnt > 0
}

func CreateTrack(ctx context.Context, track *model.Track) error {
	err := service.Create(ctx, track, service.IfNotExist())
	return err
}
