package replaygain

import (
	"context"
	"errors"
	"fmt"

	"go.senan.xyz/wrtag/tags"
)

type Addon struct {
	TruePeak bool
}

func (a Addon) ProcessRelease(ctx context.Context, paths []string) error {
	albumLev, trackLevs, err := Calculate(ctx, a.TruePeak, paths)
	if err != nil {
		return fmt.Errorf("calculate: %w", err)
	}

	var trackErrs []error
	for i := range paths {
		trackL, path := trackLevs[i], paths[i]
		err := tags.Write(path, func(f *tags.File) error {
			f.WritedB(tags.ReplayGainTrackGain, trackL.GaindB)
			f.WriteFloat(tags.ReplayGainTrackPeak, trackL.Peak)
			f.WritedB(tags.ReplayGainAlbumGain, albumLev.GaindB)
			f.WriteFloat(tags.ReplayGainAlbumPeak, albumLev.Peak)
			return nil
		})
		if err != nil {
			trackErrs = append(trackErrs, err)
			continue
		}
	}
	return errors.Join(trackErrs...)
}

func (a Addon) Name() string {
	return "replaygain"
}
