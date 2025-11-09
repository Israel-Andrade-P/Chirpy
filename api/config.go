package api

import (
	"sync/atomic"

	"github.com/Israel-Andrade-P/Chirpy.git/internal/database"
)

type Apiconfig struct {
	FileserverHits atomic.Int32
	DbQueries      *database.Queries
	Platform       string
	Secret         string
	Expiration     int
	PolkaKey       string
}
