package rpc

import (
	"context"

	pb "github.com/leebrouse/ems/backend/common/genproto/warehouse/grpc"
	"github.com/leebrouse/ems/backend/warehouse/model"
	"github.com/leebrouse/ems/backend/warehouse/service"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WarehouseRPCServer struct {
	pb.UnimplementedWarehouseServiceServer
	svc service.WarehouseService
}

func NewWarehouseRPCServer(svc service.WarehouseService) *WarehouseRPCServer {
	return &WarehouseRPCServer{svc: svc}
}

func (s *WarehouseRPCServer) ListItems(ctx context.Context, in *pb.ListItemsRequest) (*pb.ListItemsResponse, error) {
	items, total, err := s.svc.ListItems(ctx, int(in.Page), int(in.Size), in.Query)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var res []*pb.ItemResponse
	for _, item := range items {
		res = append(res, &pb.ItemResponse{
			Id:          int32(item.ID),
			Name:        item.Name,
			Unit:        item.Unit,
			Description: item.Description,
		})
	}
	return &pb.ListItemsResponse{
		Items: res,
		Total: int32(total),
	}, nil
}

func (s *WarehouseRPCServer) GetItem(ctx context.Context, in *pb.GetItemRequest) (*pb.ItemResponse, error) {
	item, err := s.svc.GetItem(ctx, int64(in.ItemId))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.ItemResponse{
		Id:          int32(item.ID),
		Name:        item.Name,
		Unit:        item.Unit,
		Description: item.Description,
	}, nil
}

func (s *WarehouseRPCServer) CreateItem(ctx context.Context, in *pb.CreateItemRequest) (*pb.ItemResponse, error) {
	item := &model.Item{
		Name:        in.Name,
		Unit:        in.Unit,
		Description: in.Description,
	}
	created, err := s.svc.CreateItem(ctx, item)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.ItemResponse{
		Id:          int32(created.ID),
		Name:        created.Name,
		Unit:        created.Unit,
		Description: created.Description,
	}, nil
}

func (s *WarehouseRPCServer) UpdateItem(ctx context.Context, in *pb.UpdateItemRequest) (*pb.ItemResponse, error) {
	updated, err := s.svc.UpdateItem(ctx, int64(in.ItemId), in.Name, in.Unit, in.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.ItemResponse{
		Id:          int32(updated.ID),
		Name:        updated.Name,
		Unit:        updated.Unit,
		Description: updated.Description,
	}, nil
}

// 
func (s *WarehouseRPCServer) DeleteItem(ctx context.Context, in *pb.DeleteItemRequest) (*empty.Empty, error) {
	err := s.svc.DeleteItem(ctx, int64(in.ItemId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

// 
func (s *WarehouseRPCServer) ListWarehouses(ctx context.Context, in *pb.ListWarehousesRequest) (*pb.ListWarehousesResponse, error) {
	ws, err := s.svc.ListWarehouses(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var res []*pb.WarehouseResponse
	for _, w := range ws {
		res = append(res, &pb.WarehouseResponse{
			Id:       int32(w.ID),
			Name:     w.Name,
			Location: w.Location,
		})
	}
	return &pb.ListWarehousesResponse{Warehouses: res}, nil
}

// 
func (s *WarehouseRPCServer) GetWarehouse(ctx context.Context, in *pb.GetWarehouseRequest) (*pb.WarehouseResponse, error) {
	w, err := s.svc.GetWarehouse(ctx, int64(in.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.WarehouseResponse{
		Id:       int32(w.ID),
		Name:     w.Name,
		Location: w.Location,
	}, nil
}

func (s *WarehouseRPCServer) CreateWarehouse(ctx context.Context, in *pb.CreateWarehouseRequest) (*pb.WarehouseResponse, error) {
	w, err := s.svc.CreateWarehouse(ctx, in.Name, in.Location)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.WarehouseResponse{
		Id:       int32(w.ID),
		Name:     w.Name,
		Location: w.Location,
	}, nil
}

func (s *WarehouseRPCServer) UpdateWarehouse(ctx context.Context, in *pb.UpdateWarehouseRequest) (*pb.WarehouseResponse, error) {
	w, err := s.svc.UpdateWarehouse(ctx, int64(in.Id), in.Name, in.Location)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.WarehouseResponse{
		Id:       int32(w.ID),
		Name:     w.Name,
		Location: w.Location,
	}, nil
}

func (s *WarehouseRPCServer) DeleteWarehouse(ctx context.Context, in *pb.DeleteWarehouseRequest) (*empty.Empty, error) {
	err := s.svc.DeleteWarehouse(ctx, int64(in.Id))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func (s *WarehouseRPCServer) AdjustInventory(ctx context.Context, in *pb.AdjustInventoryRequest) (*pb.InventoryResponse, error) {
	// For gRPC AdjustInventory, we assume it might come from other services (like scheduling)
	// We'll use "gRPC" as reference type for now.
	inv, err := s.svc.AdjustInventory(ctx, int64(in.WarehouseId), int64(in.ItemId), int(in.Amount), "GRPC", 0)
	if err != nil {
		if err == service.ErrInsufficientStock {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.InventoryResponse{
		ItemId:   int32(inv.ItemID),
		Quantity: int32(inv.Quantity),
	}, nil
}
