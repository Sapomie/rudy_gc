package modelx

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func (m *defaultAMovieModel) FindRandomMovies(ctx context.Context, count int) ([]*AMovie, error) {
	// Step 1: Get the total number of movies
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, err
	}

	// Step 2: Generate unique random indices
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	indexMap := make(map[int64]struct{}, count)
	for len(indexMap) < count && len(indexMap) < int(total) {
		randomID := int64(rng.Intn(int(total))) + 1 // Assuming IDs are 1-indexed
		indexMap[randomID] = struct{}{}
	}

	// Step 3: Fetch movies using random IDs
	movies := make([]*AMovie, 0, count)
	for id := range indexMap {
		movie, err := m.FindOne(ctx, id)
		if err != nil {
			// Here we assume if a movie does not exist treat it as a legitimate absence
			if err != sqlx.ErrNotFound {
				return nil, fmt.Errorf("error fetching movie with id %d: %w", id, err)
			}
			continue
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

func (m *defaultAMovieModel) FindRandomMoviesOwn(ctx context.Context, count int) ([]*AMovie, error) {
	// Step 1: Fetch all IDs where film_birth_time > 0
	var ids []int64
	query := fmt.Sprintf("SELECT `id` FROM %s WHERE `movie_owned` in (2,3) ", m.table)

	err := m.QueryRowsNoCacheCtx(ctx, &ids, query)
	if err != nil {
		return nil, err
	}

	// If the requested count is greater than available, adjust the count
	if count > len(ids) {
		count = len(ids)
	}

	// Step 2: Shuffle the slice of IDs and pick the first 'count' IDs
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(ids), func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})

	selectedIDs := ids[:count]

	// Step 3: Fetch movies using the selected random IDs via FindOne
	movies := make([]*AMovie, 0, count)
	for _, id := range selectedIDs {
		movie, err := m.FindOne(ctx, id)
		if err != nil {
			if err != sqlx.ErrNotFound {
				return nil, fmt.Errorf("error fetching movie with id %d: %w", id, err)
			}
			continue
		}

		// Ensure the movie satisfies the condition
		if movie.FilmBirthTime > 0 {
			movies = append(movies, movie)
		}

		// Check if we have gathered enough movies
		if len(movies) == count {
			break
		}
	}

	return movies, nil
}
