package gRPC_server

import (
	"context"
	"log"
	"net"
	pb "github.com/SyntinelNyx/syntinel-server/internal/gRPC_server/internal/hw_info"
	"google.golang.org/grpc"
)

func (s *server) GetInfo(stream pb.RouteGuide_RecordRouteServer) error {
	var pointCount, featureCount, distance int32
	var lastPoint *pb.Point
	startTime := time.Now()
	for {
	  point, err := stream.Recv()
	  if err == io.EOF {
		endTime := time.Now()
		return stream.SendAndClose(&pb.RouteSummary{
		  PointCount:   pointCount,
		  FeatureCount: featureCount,
		  Distance:     distance,
		  ElapsedTime:  int32(endTime.Sub(startTime).Seconds()),
		})
	  }
	  if err != nil {
		return err
	  }
	  pointCount++
	  for _, feature := range s.savedFeatures {
		if proto.Equal(feature.Location, point) {
		  featureCount++
		}
	  }
	  if lastPoint != nil {
		distance += calcDistance(lastPoint, point)
	  }
	  lastPoint = point
	}
  }
