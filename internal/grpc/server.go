// Package grpc provides gRPC server implementation for the shortener service.
// It exposes the same business logic as the HTTP handlers through gRPC,
// including URL shortening, expansion, and user URL listing with JWT authentication
// via metadata.
package grpc

import (
	"context"
	"errors"
	"strings"

	"shorturl/internal/config"
	appctx "shorturl/internal/context"
	"shorturl/internal/logger"
	"shorturl/internal/middleware"
	"shorturl/internal/service"
	pb "shorturl/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server implements pb.ShortenerServiceServer and delegates to the shared business logic.
type Server struct {
	pb.UnimplementedShortenerServiceServer
	linkService service.LinkServiceInterface
	cfg         config.Config
	logger      *zap.Logger
}

// NewServer creates a new gRPC server instance.
func NewServer(linkService service.LinkServiceInterface, cfg config.Config, log *zap.Logger) *Server {
	return &Server{
		linkService: linkService,
		cfg:         cfg,
		logger:      log,
	}
}

// ShortenURL handles the ShortenURL RPC - equivalent to POST /api/shorten.
func (s *Server) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	userID, ok := appctx.UserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	url := strings.TrimSpace(req.GetUrl())
	if url == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	id, err := s.linkService.Create(url, userID)
	if err != nil {
		var conflictErr *service.URLAlreadyExistsError
		if errors.As(err, &conflictErr) {
			result := s.cfg.BaseURL + "/" + conflictErr.ID
			return &pb.URLShortenResponse{
				Result: &result,
			}, nil
		}

		s.logger.Error("error creating short url", zap.Error(err))
		return nil, status.Error(codes.Internal, "error while creating")
	}

	result := s.cfg.BaseURL + "/" + id
	return &pb.URLShortenResponse{
		Result: &result,
	}, nil
}

// ExpandURL handles the ExpandURL RPC - equivalent to GET /{id}.
func (s *Server) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	id := strings.TrimSpace(req.GetId())
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	url, err := s.linkService.Get(id)
	if err != nil {
		if errors.Is(err, service.ErrLinkDeleted) {
			return nil, status.Error(codes.NotFound, "link deleted")
		}
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "link not found")
		}

		s.logger.Error("error getting short url", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.URLExpandResponse{
		Result: &url,
	}, nil
}

// ListUserURLs handles the ListUserURLs RPC - equivalent to GET /api/user/urls.
func (s *Server) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, ok := appctx.UserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	links, err := s.linkService.GetAllByUserID(userID)
	if err != nil {
		s.logger.Error("error listing user urls", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	urls := make([]*pb.URLData, 0, len(links))
	for _, link := range links {
		shortURL := s.cfg.BaseURL + "/" + link.ID
		origURL := link.URL
		urls = append(urls, &pb.URLData{
			ShortUrl:    &shortURL,
			OriginalUrl: &origURL,
		})
	}

	return &pb.UserURLsResponse{
		Url: urls,
	}, nil
}

// AuthInterceptor is a gRPC unary server interceptor that extracts JWT from
// the "authorization" metadata header for authentication.
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		var tokenString string

		// Try "authorization" first, then "token" header
		if values := md.Get("authorization"); len(values) > 0 {
			tokenString = values[0]
		} else if values := md.Get("token"); len(values) > 0 {
			tokenString = values[0]
		}

		if tokenString == "" {
			// Generate a new token for new users (same behavior as HTTP cookie middleware)
			var buildErr error
			tokenString, buildErr = middleware.BuildJWTString(logger.Log)
			if buildErr != nil {
				logger.Log.Error("failed to build jwt token", zap.Error(buildErr))
				return nil, status.Error(codes.Internal, "failed to build jwt token")
			}
		}

		userID, err := middleware.GetUserID(tokenString)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = appctx.WithUserID(ctx, userID)
		return handler(ctx, req)
	}
}
