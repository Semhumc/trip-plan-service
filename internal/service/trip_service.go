// internal/service/trip_service.go - Eksik metodların eklenmesi

package service

import (
	"context"
	"database/sql"
	"time"
	db "trip-plan-service/internal/db/postgresql"
	"trip-plan-service/internal/models"
)

type TripService struct {
	TripSer   *models.Trip
	DB        *sql.DB
	Locations []models.Location
	Queries   *db.Queries
}

func NewTripService(trip *models.Trip, dbConn *sql.DB, locations []models.Location) *TripService {
	return &TripService{
		TripSer:   trip,
		DB:        dbConn,
		Locations: locations,
		Queries:   db.New(dbConn),
	}
}

func (s *TripService) SaveTripWLocations(ctx context.Context) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	qtx := s.Queries.WithTx(tx)

	trip, err := qtx.Create_Trip(ctx, db.Create_TripParams{
		UserID:      s.TripSer.UserID,
		Name:        s.TripSer.Name,
		Description: sql.NullString{
			String: s.TripSer.Description,
			Valid:  s.TripSer.Description != "",
		},
		StartDate: s.TripSer.StartDate,
		EndDate:   s.TripSer.EndDate,
	})

	if err != nil {
		tx.Rollback()
		return err
	}

	for i, loc := range s.Locations {
		location, err := qtx.Create_Location(ctx, db.Create_LocationParams{
			Name: loc.Name,
			Address: sql.NullString{
				String: func() string {
					if loc.Address != nil {
						return *loc.Address
					}
					return ""
				}(),
				Valid: loc.Address != nil && *loc.Address != "",
			},
			SiteUrl: sql.NullString{
				String: func() string {
					if loc.SiteURL != nil {
						return *loc.SiteURL
					}
					return ""
				}(),
				Valid: loc.SiteURL != nil && *loc.SiteURL != "",
			},
			Notes: sql.NullString{
				String: func() string {
					if loc.Notes != nil {
						return *loc.Notes
					}
					return ""
				}(),
				Valid: loc.Notes != nil && *loc.Notes != "",
			},
		})
		if err != nil {
			tx.Rollback()
			return err
		}

		err = qtx.Create_Trip_Location(ctx, db.Create_Trip_LocationParams{
			TripID:     int32(trip.ID),
			LocationID: int32(location.ID),
			Position:   int32(i + 1),
		})
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// YENİ: Kullanıcının tüm triplerini getir
func (s *TripService) GetUserTrips(ctx context.Context, userID string) ([]models.TripWithLocations, error) {
	trips, err := s.Queries.ListTripsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []models.TripWithLocations
	for _, trip := range trips {
		// Her trip için location'ları getir
		locations, err := s.Queries.GetTripLocations(ctx, trip.ID)
		if err != nil {
			return nil, err
		}

		var tripLocations []models.Location
		for _, loc := range locations {
			tripLocation := models.Location{
				ID:   int(loc.ID),
				Name: loc.Name,
				Address: func() *string {
					if loc.Address.Valid {
						return &loc.Address.String
					}
					return nil
				}(),
				SiteURL: func() *string {
					if loc.SiteUrl.Valid {
						return &loc.SiteUrl.String
					}
					return nil
				}(),
				Notes: func() *string {
					if loc.Notes.Valid {
						return &loc.Notes.String
					}
					return nil
				}(),
				CreatedAt: func() time.Time {
					if loc.CreatedAt.Valid {
						return loc.CreatedAt.Time
					}
					return time.Time{}
				}(),
			}
			tripLocations = append(tripLocations, tripLocation)
		}

		tripWithLoc := models.TripWithLocations{
			Trip: models.Trip{
				ID:          int(trip.ID),
				UserID:      trip.UserID,
				Name:        trip.Name,
				Description: func() string {
					if trip.Description.Valid {
						return trip.Description.String
					}
					return ""
				}(),
				StartDate: trip.StartDate,
				EndDate:   trip.EndDate,
				CreatedAt: func() time.Time {
					if trip.CreatedAt.Valid {
						return trip.CreatedAt.Time
					}
					return time.Time{}
				}(),
				UpdatedAt: func() time.Time {
					if trip.UpdatedAt.Valid {
						return trip.UpdatedAt.Time
					}
					return time.Time{}
				}(),
				StartPosition: func() string {
					if trip.StartPosition.Valid {
						return trip.StartPosition.String
					}
					return ""
				}(),
				EndPosition: func() string {
					if trip.FinishPosition.Valid {
						return trip.FinishPosition.String
					}
					return ""
				}(),
			},
			Locations: tripLocations,
		}
		result = append(result, tripWithLoc)
	}

	return result, nil
}

// YENİ: Trip sil
func (s *TripService) DeleteTrip(ctx context.Context, tripID int32) error {
	return s.Queries.DeleteTrip(ctx, tripID)
}

// YENİ: ID'ye göre trip getir
func (s *TripService) GetTripByID(ctx context.Context, tripID int32) (*models.TripWithLocations, error) {
	trip, err := s.Queries.GetTripByID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	locations, err := s.Queries.GetTripLocations(ctx, tripID)
	if err != nil {
		return nil, err
	}

	var tripLocations []models.Location
	for _, loc := range locations {
		tripLocation := models.Location{
			ID:   int(loc.ID),
			Name: loc.Name,
			Address: func() *string {
				if loc.Address.Valid {
					return &loc.Address.String
				}
				return nil
			}(),
			SiteURL: func() *string {
				if loc.SiteUrl.Valid {
					return &loc.SiteUrl.String
				}
				return nil
			}(),
			Notes: func() *string {
				if loc.Notes.Valid {
					return &loc.Notes.String
				}
				return nil
			}(),
			CreatedAt: func() time.Time {
				if loc.CreatedAt.Valid {
					return loc.CreatedAt.Time
				}
				return time.Time{}
			}(),
		}
		tripLocations = append(tripLocations, tripLocation)
	}

	result := &models.TripWithLocations{
		Trip: models.Trip{
			ID:          int(trip.ID),
			UserID:      trip.UserID,
			Name:        trip.Name,
			Description: func() string {
				if trip.Description.Valid {
					return trip.Description.String
				}
				return ""
			}(),
			StartDate: trip.StartDate,
			EndDate:   trip.EndDate,
			CreatedAt: func() time.Time {
				if trip.CreatedAt.Valid {
					return trip.CreatedAt.Time
				}
				return time.Time{}
			}(),
			UpdatedAt: func() time.Time {
				if trip.UpdatedAt.Valid {
					return trip.UpdatedAt.Time
				}
				return time.Time{}
			}(),
			StartPosition: func() string {
				if trip.StartPosition.Valid {
					return trip.StartPosition.String
				}
				return ""
			}(),
			EndPosition: func() string {
				if trip.FinishPosition.Valid {
					return trip.FinishPosition.String
				}
				return ""
			}(),
		},
		Locations: tripLocations,
	}

	return result, nil
}


// YENİ: Kullanıcının tüm triplerini getir
