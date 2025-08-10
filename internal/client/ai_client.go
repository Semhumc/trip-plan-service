// internal/client/ai_client.go for trip-plan-service
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Semhumc/grpc-proto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AIClient struct {
	client proto.AIServiceClient
	conn   *grpc.ClientConn
}

func NewAIClient(serverAddress string) (*AIClient, error) {
	conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AI service: %v", err)
	}

	client := proto.NewAIServiceClient(conn)
	
	return &AIClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *AIClient) Close() error {
	return c.conn.Close()
}

func (c *AIClient) GenerateTripPlan(ctx context.Context, req *proto.PromptRequest) (*proto.TripOptionsResponse, error) {
	// Set timeout for the request
	ctx, cancel := context.WithTimeout(ctx, 250*time.Second)
	defer cancel()

	response, err := c.client.GeneratePlan(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate trip plan: %v", err)
	}

	return response, nil
}

// Helper function to convert internal models to proto
func CreatePromptRequest(userID, name, description, startPos, endPos, startDate, endDate string) *proto.PromptRequest {
	return &proto.PromptRequest{
		UserId:        userID,
		Name:          name,
		Description:   description,
		StartPosition: startPos,
		EndPosition:   endPos,
		StartDate:     startDate,
		EndDate:       endDate,
	}
}