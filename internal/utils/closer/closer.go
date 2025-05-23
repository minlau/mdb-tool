package closer

import (
	"io"

	"github.com/rs/zerolog/log"
)

// Handle calls Close() and logs error
func Handle(closer io.Closer, msg string) {
	err := closer.Close()
	if err != nil {
		log.Error().Err(err).Msgf("failed to close %s", msg)
	}
}
