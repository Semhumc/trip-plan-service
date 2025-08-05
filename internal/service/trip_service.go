package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
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
	// Hata durumunda Rollback'i garantilemek için defer kullanın.
	defer tx.Rollback()

	qtx := s.Queries.WithTx(tx)

	trip, err := qtx.CreateTrip(ctx, db.CreateTripParams{
		UserID: s.TripSer.UserID,
		Name:   s.TripSer.Name,
		StartPosition: sql.NullString{
			String: s.TripSer.StartPosition,
			Valid:  s.TripSer.StartPosition != "",
		},
		EndPosition: sql.NullString{
			String: s.TripSer.EndPosition,
			Valid:  s.TripSer.EndPosition != "",
		},
		Description: sql.NullString{
			String: s.TripSer.Description,
			Valid:  s.TripSer.Description != "",
		},
		StartDate: func() time.Time {
			t, _ := time.Parse("2006-01-02", s.TripSer.StartDate)
			return t
		}(),
		EndDate: func() time.Time {
			t, _ := time.Parse("2006-01-02", s.TripSer.EndDate)
			return t
		}(),
	})

	if err != nil {
		return err // Rollback defer ile yapılacak
	}

	for i, loc := range s.Locations {
		location, err := qtx.CreateLocation(ctx, db.CreateLocationParams{
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
			Latitude: sql.NullString{
				String: fmt.Sprintf("%f", loc.Latitude),
				Valid:  true,
			},
			// EKLENDİ: Eksik olan Longitude parametresi eklendi.
			Longitude: sql.NullString{
				String: fmt.Sprintf("%f", loc.Longitude),
				Valid:  true,
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
			return err // Rollback defer ile yapılacak
		}

		// DÜZELTİLDİ: Fonksiyon adı `CreateTripLocation`'dan `AddLocationToTrip`'e çevrildi.
		err = qtx.AddLocationToTrip(ctx, db.AddLocationToTripParams{
			TripID:     trip.ID,
			LocationID: location.ID,
			Position:   int32(i + 1),
		})
		if err != nil {
			return err // Rollback defer ile yapılacak
		}
	}

	// Her şey başarılıysa Commit et
	return tx.Commit()
}

func (s *TripService) GetUserTrips(ctx context.Context, userID string) ([]models.TripWithLocations, error) {
	trips, err := s.Queries.ListTripsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []models.TripWithLocations
	for _, trip := range trips {
		locationsDB, err := s.Queries.GetTripLocations(ctx, trip.ID)
		if err != nil {
			return nil, err
		}

		var tripLocations []models.Location
		for _, loc := range locationsDB {
			tripLocation := models.Location{
				ID:   int(loc.ID),
				Name: loc.Name,
				Latitude: func() float64 {
					if loc.Latitude.Valid {
						val, err := strconv.ParseFloat(loc.Latitude.String, 64)
						if err == nil {
							return val
						}
					}
					return 0
				}(),
				Longitude: func() float64 {
					if loc.Longitude.Valid {
						val, err := strconv.ParseFloat(loc.Longitude.String, 64)
						if err == nil {
							return val
						}
					}
					return 0
				}(),
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
				CreatedAt: loc.CreatedAt.Time,
			}
			tripLocations = append(tripLocations, tripLocation)
		}

		tripWithLoc := models.TripWithLocations{
			Trip: models.Trip{
				ID:            int(trip.ID),
				UserID:        trip.UserID,
				Name:          trip.Name,
				Description:   trip.Description.String,
				StartDate:     trip.StartDate.Format("2006-01-02"),
				EndDate:       trip.EndDate.Format("2006-01-02"),
				CreatedAt:     trip.CreatedAt.Time,
				UpdatedAt:     trip.UpdatedAt.Time,
				StartPosition: trip.StartPosition.String,
				// DÜZELTİLDİ: SQLC artık EndPosition olarak üretecek
				EndPosition: trip.EndPosition.String,
			},
			Locations: tripLocations,
		}
		result = append(result, tripWithLoc)
	}

	return result, nil
}

func (s *TripService) DeleteTrip(ctx context.Context, tripID int32) error {
	return s.Queries.DeleteTrip(ctx, tripID)
}

func (s *TripService) GetTripByID(ctx context.Context, tripID int32) (*models.TripWithLocations, error) {
	trip, err := s.Queries.GetTripByID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	locationsDB, err := s.Queries.GetTripLocations(ctx, tripID)
	if err != nil {
		return nil, err
	}

	var tripLocations []models.Location
	for _, loc := range locationsDB {
		tripLocation := models.Location{
			ID:   int(loc.ID),
			Name: loc.Name,
			Latitude: func() float64 {
				if loc.Latitude.Valid {
					val, err := strconv.ParseFloat(loc.Latitude.String, 64)
					if err == nil {
						return val
					}
				}
				return 0
			}(),
			Longitude: func() float64 {
				if loc.Longitude.Valid {
					val, err := strconv.ParseFloat(loc.Longitude.String, 64)
					if err == nil {
						return val
					}
				}
				return 0
			}(),
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
			CreatedAt: loc.CreatedAt.Time,
		}
		tripLocations = append(tripLocations, tripLocation)
	}

	result := &models.TripWithLocations{
		Trip: models.Trip{
			ID:            int(trip.ID),
			UserID:        trip.UserID,
			Name:          trip.Name,
			Description:   trip.Description.String,
			StartDate:     trip.StartDate.Format("2006-01-02"),
			EndDate:       trip.EndDate.Format("2006-01-02"),
			CreatedAt:     trip.CreatedAt.Time,
			UpdatedAt:     trip.UpdatedAt.Time,
			StartPosition: trip.StartPosition.String,
			// DÜZELTİLDİ: SQLC artık EndPosition olarak üretecek
			EndPosition: trip.EndPosition.String,
		},
		Locations: tripLocations,
	}

	return result, nil
}
