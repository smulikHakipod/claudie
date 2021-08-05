package terraformer

import (
	"context"
	"fmt"
	"log"

	"github.com/Berops/platform/proto/pb"
)

func BuildInfrastructure(c pb.TerraformerServiceClient, req *pb.BuildInfrastructureRequest) (*pb.BuildInfrastructureResponse, error) {
	res, err := c.BuildInfrastructure(context.Background(), req) //sending request to the server and receiving response
	if err != nil {
		return nil, fmt.Errorf("error while calling BuildInfrastructure on Terraformer: %v", err)
	}

	log.Println("Infrastructure was successfully built")
	return res, nil
}

func DestroyInfrastructure(c pb.TerraformerServiceClient, req *pb.DestroyInfrastructureRequest) (*pb.DestroyInfrastructureResponse, error) {
	res, err := c.DestroyInfrastructure(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("error while calling DestroyInfrastructure on Terraformer: %v", err)
	}

	log.Println("Infrastructure was successfully destroyed")
	return res, nil
}
