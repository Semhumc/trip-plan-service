package service

import (
	"context"
	"database/sql"
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
		StartDate:   s.TripSer.StartDate,
		EndDate:     s.TripSer.EndDate,
	})

	if err !=nil{
		tx.Rollback()
		return err
	} 

	for i, loc :=range s.Locations{
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
