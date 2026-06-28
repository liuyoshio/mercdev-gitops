package server

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/liuyoshio/platformd/internal/catalog"
	catalogv1 "github.com/liuyoshio/platformd/proto/catalogv1"
)

// CatalogServer adapts the catalog.Store to the gRPC interface.
type CatalogServer struct {
	catalogv1.UnimplementedCatalogServiceServer // forward-compat embedding
	store                                       *catalog.Store
}

func NewCatalogServer(store *catalog.Store) *CatalogServer {
	return &CatalogServer{store: store}
}

func (s *CatalogServer) RegisterService(ctx context.Context, req *catalogv1.RegisterServiceRequest) (*catalogv1.RegisterServiceResponse, error) {
	ctx, span := otel.Tracer("platformd").Start(ctx, "catalog.Register")
	defer span.End()

	in := req.GetService()
	span.SetAttributes(attribute.String("catalog.service.name", in.GetName()))

	if in.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	svc, err := s.store.Register(catalog.Service{
		Name:     in.GetName(),
		Owner:    in.GetOwner(),
		Language: in.GetLanguage(),
		Replicas: in.GetReplicas(),
	})
	if err == catalog.ErrExists {
		return nil, status.Errorf(codes.AlreadyExists, "%s already registered", in.GetName())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &catalogv1.RegisterServiceResponse{Service: toProto(svc)}, nil
}

func (s *CatalogServer) GetService(ctx context.Context, req *catalogv1.GetServiceRequest) (*catalogv1.GetServiceResponse, error) {
	svc, err := s.store.Get(req.GetName())
	if err == catalog.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "%s not found", req.GetName())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &catalogv1.GetServiceResponse{Service: toProto(svc)}, nil
}

func (s *CatalogServer) ListServices(ctx context.Context, _ *catalogv1.ListServicesRequest) (*catalogv1.ListServicesResponse, error) {
	all := s.store.List()
	out := make([]*catalogv1.Service, 0, len(all))
	for _, svc := range all {
		out = append(out, toProto(svc))
	}
	return &catalogv1.ListServicesResponse{Services: out}, nil
}

// toProto converts the domain model to the wire type.
func toProto(s catalog.Service) *catalogv1.Service {
	return &catalogv1.Service{
		Name:     s.Name,
		Owner:    s.Owner,
		Language: s.Language,
		Replicas: s.Replicas,
	}
}
