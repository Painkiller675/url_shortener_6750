package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/controller"
	"github.com/Painkiller675/url_shortener_6750/internal/lib/merrors"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
	"github.com/Painkiller675/url_shortener_6750/internal/protos"
	"github.com/Painkiller675/url_shortener_6750/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"net/url"
	"sync"
)

// Server - the router(like REST controller here) and the thingy that provides us with some handlers
type Server struct {
	business                            controller.Business
	logger                              *zap.Logger
	wg                                  *sync.WaitGroup
	delJobs                             chan models.JobToDelete
	protos.UnimplementedShortenerServer // it calls that we use gRPC here
}

func NewGRPCServer(business controller.Business, logger *zap.Logger, wg *sync.WaitGroup, delJobs chan models.JobToDelete) *Server {
	return &Server{business: business, logger: logger, wg: wg, delJobs: delJobs}
}

// PingDB matches "/ping", c.PingDB() in the REST controller
func (s *Server) PingDB(ctx context.Context, _ *protos.PingDBRequest) (*protos.PingDBResponse, error) {
	if err := s.business.PingDB(ctx); err != nil {
		return nil, fmt.Errorf("err: %w", err)
	}
	fmt.Println("gRPC ping is detected")
	return &protos.PingDBResponse{}, nil
}

// CreateShortURL matches "/", c.CreateShortURLHandler()
func (s *Server) CreateShortURL(ctx context.Context, in *protos.CreateShortURLRequest) (*protos.CreateShortURLResponse, error) {
	const op = "grpc.CreateShortURL"
	// get some metadata and retrieve an authorization token if any
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
	token := md.Get("token") // returns an array of values for the key
	if len(token) == 0 || token[0] == "" {
		return nil, status.Error(codes.Unauthenticated, "missing the authorization token")
	}
	// retrieve userID if any
	userID, err := s.business.RetrieveUserIDFromTokenString(token[0])
	if err != nil { // can't retrieve => register a new user a
		tokenStr, _, err := s.business.GenJWTTokenString()
		if err != nil {
			s.logger.Info("Can't generate token!", zap.Error(err))
			return nil, status.Error(codes.Internal, "can't generate the token")

		}
		// set generate token string into the header for a new user (in REST we set it into the cookies)
		if err := grpc.SetHeader(ctx, metadata.Pairs("token", tokenStr)); err != nil {
			s.logger.Info("Can't send token!", zap.Error(err))
			return nil, status.Error(codes.Internal, "can't send the token")
		}
	}
	// parse the url (validation +)
	urlIn, err := url.Parse(in.Url)
	if err != nil {
		s.logger.Error("[ERROR] can't parse a client URL", zap.String("place:", op), zap.Error(err))
		return nil, status.Error(codes.Internal, "can't parse a client URL")
	}
	// save the data
	resultURL, err := s.business.StoreAlURL(ctx, urlIn, userID) // TODO: mb del _ or change driver to use id?
	s.logger.Info("[INSERT] a try", zap.String("ShortURL:", resultURL), zap.String("InputURL: ", in.Url))
	if err != nil {
		if errors.Is(err, merrors.ErrURLOrAliasExists) { // the try to short already existed url pg database
			s.logger.Info("URL already exists!", zap.String("place:", op), zap.Error(err))
			// response molding (existing URL && 409(REST))
			return nil, status.Error(codes.AlreadyExists, "URL already exists")

		} else {
			s.logger.Info("Failed to store URL", zap.String("place:", op), zap.Error(err))
			return nil, status.Error(codes.Internal, "can't store URL")
		}

	}
	// if everything is ok => add url into the database
	// response molding

	return &protos.CreateShortURLResponse{Result: resultURL}, nil
}

// Delete matches "/api/user/urls", c.DeleteURLSHandler() in the REST controller
func (s *Server) Delete(ctx context.Context, in *protos.DeleteRequest) (*protos.DeleteResponse, error) {
	const op = "grpc.Delete"
	// get some metadata and retrieve an authorization token if any
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
	token := md.Get("token") // returns an array of values for the key
	if len(token) == 0 || token[0] == "" {
		return nil, status.Error(codes.Unauthenticated, "missing the authorization token")
	}
	// retrieve userID if any
	userID, err := s.business.RetrieveUserIDFromTokenString(token[0])
	if err != nil {
		s.logger.Info("can't retrieve userID", zap.String("place:", op), zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "can't retrieve userID")
	}
	// check if user exists TODO: del that or not
	err = s.business.CheckIfUserExists(ctx, userID)
	if err != nil {
		if errors.Is(err, merrors.ErrUserNotFound) {
			s.logger.Info("User not found", zap.String("place:", op), zap.String("user_id", userID))
			return nil, status.Error(codes.NotFound, "user not found")
		}
		// handle other possible errors (unexpected ones)
		s.logger.Error("Failed to check if user exists", zap.String("place:", op), zap.Error(err))
		return nil, status.Error(codes.Internal, "can't check if user exists")
	}
	// deleting
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.delJobs <- models.JobToDelete{UserID: userID, LsURL: in.Ids}
	}()
	return &protos.DeleteResponse{}, nil
}

// GetURL matches "/{id}", c.GetLongURLHandler in the REST controller
func (s *Server) GetURL(ctx context.Context, in *protos.GetUrlRequest) (*protos.GetUrlResponse, error) {
	// get the alias (id) from the request
	const op = "grpc.GetURL"
	id := in.GetId()
	if id == "" {
		s.logger.Info("request id is empty", zap.String("place:", op))
		return nil, status.Error(codes.InvalidArgument, "id is empty")
	}
	orURL, err := s.business.GetOrURLByAl(ctx, id)
	if err != nil { // TODO: mb I should use status 500 here?
		if errors.Is(err, merrors.ErrURLIsDel) { // if URL was deleted
			s.logger.Info("[INFO]", zap.String("place:", op), zap.Error(err))
			return nil, status.Error(codes.NotFound, "url not found")
		}
		// OTHER ERRORS
		s.logger.Info("Failed to get orURL", zap.String("place:", op), zap.String("id", id), zap.Error(err))
		return nil, status.Error(codes.Internal, "can't get orURL")
	}
	return &protos.GetUrlResponse{Url: orURL}, nil
}

// Batch matches "/api/shorten/batch", c.CreateShortURLJSONBatchHandler() in the REST controller
func (s *Server) Batch(ctx context.Context, in *protos.BatchRequest) (*protos.BatchResponse, error) {
	const op = "grpc.Batch"
	// create the array of structures to feed it into s.business.SaveBatchURL
	var desBatchStruct []models.JSONBatStructToDesReq

	for _, pbURL := range in.Urls {
		desBatchStruct = append(desBatchStruct, models.JSONBatStructToDesReq{CorrelationID: pbURL.GetCorrelationId(), OriginalURL: pbURL.GetOriginalUrl()})
	}
	// create an auxiliary array of structures
	idURLAl, err := service.CreateBatchIDOrSh(&desBatchStruct)
	if err != nil {
		s.logger.Error("[ERROR]", zap.String("place:", op), zap.Error(err))
		return nil, status.Error(codes.Internal, "can't transform into a proxy-struct")
	}
	// save data into the database
	respBatch, err := s.business.SaveBatchURL(ctx, idURLAl)
	if err != nil {
		s.logger.Error("[ERROR]", zap.String("place:", op), zap.Error(err))
		return nil, status.Error(codes.Internal, "can't save the batch of URLs")
	}
	// mold the pb-response
	var pbBatchOfURLsResp []*protos.BatchShortUrl
	for _, batch := range *respBatch { // TODO [MENTOR]: iteration by * or without it??
		pbBatchOfURLsResp = append(pbBatchOfURLsResp, &protos.BatchShortUrl{CorrelationId: batch.CorrelationID, ShortUrl: batch.ShortURL})
	}
	return &protos.BatchResponse{Result: pbBatchOfURLsResp}, nil
}

func (s *Server) GetURLs(ctx context.Context, in *protos.GetUrlsRequest) (*protos.GetUrlsResponse, error) {
	const op = "grpc.GetURLs"
	// get some metadata and retrieve an authorization token if any
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
	token := md.Get("token") // returns an array of values for the key
	if len(token) == 0 || token[0] == "" {
		return nil, status.Error(codes.Unauthenticated, "missing the authorization token")
	}
	// retrieve userID if any
	userID, err := s.business.RetrieveUserIDFromTokenString(token[0])
	if err != nil { // can't retrieve => register a new user a
		tokenStr, _, err := s.business.GenJWTTokenString()
		if err != nil {
			s.logger.Info("Can't generate token!", zap.Error(err))
			return nil, status.Error(codes.Internal, "can't generate the token")

		}
		// set generate token string into the header for a new user (in REST we set it into the cookies)
		if err := grpc.SetHeader(ctx, metadata.Pairs("token", tokenStr)); err != nil {
			s.logger.Info("Can't send token!", zap.Error(err))
			return nil, status.Error(codes.Internal, "can't send the token")
		}
	}
	// get the all user's aliases by his id
	respAlURLStruct, err := s.business.GetDataByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, merrors.ErrURLNotFound) { // no data for the user!
			s.logger.Info("[INFO]", zap.String("place:", op), zap.Error(err))
			return nil, status.Error(codes.NotFound, "url not found")
		}
		// handle other possible errors
		s.logger.Error("[ERROR]", zap.String("place:", op), zap.Error(err))
		return nil, status.Error(codes.Internal, "can't handle")
	}
	// replace alias with short url (add base url)
	for n, alURL := range *respAlURLStruct {
		fullShortURL := (config.StartOptions.BaseURL.JoinPath(alURL.ShortURL)).String()
		(*(respAlURLStruct))[n].ShortURL = fullShortURL
	}
	// create the slice of pb-type
	var pbUrls []*protos.ShortUrl
	// fill the pb-type array to set it into the request then
	for _, value := range *respAlURLStruct {
		pbUrls = append(pbUrls, &protos.ShortUrl{
			ShortUrl:    value.ShortURL,
			OriginalUrl: value.OriginalURL,
		})
	}

	return &protos.GetUrlsResponse{
		Urls: pbUrls,
	}, nil

}

// InternalStats matches "/api/internal/stats", c.GetStats()
func (s *Server) InternalStats(ctx context.Context, _ *protos.StatRequest) (*protos.StatResponse, error) {
	const op = "grpc.InternalStats"

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	clientIPStr := md.Get("X-Real-IP")
	if len(clientIPStr) == 0 || clientIPStr[0] == "" {
		return nil, status.Error(codes.Unauthenticated, "missing X-Real-IP")
	}

	clientIP := net.ParseIP(clientIPStr[0])

	// get the subnet from config/flag/env
	_, trustedNet, err := net.ParseCIDR(config.StartOptions.TrustedSubnet)
	if err != nil {
		s.logger.Info("Invalid Trusted Subnet", zap.String("place:", op), zap.String("subnet", config.StartOptions.TrustedSubnet))
		return nil, status.Error(codes.Internal, "Invalid config Subnet")
	}
	// check if permission denied or not
	if trustedNet == nil || clientIP == nil || !trustedNet.Contains(clientIP) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}
	// get needed statistics from the database
	urlsCount, usersCount, err := s.business.GetStats(ctx)
	if err != nil {
		s.logger.Info("Error getting stats", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get count urls")
	}
	return &protos.StatResponse{
		Urls:  int32(urlsCount),
		Users: int32(usersCount),
	}, nil
}

// Serve - relates grpcServer with Shorten service  and launches the gRPC server
func Serve(srv *Server) error {
	lis, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		return fmt.Errorf("failed to run gRPC server: %v", err)
	}
	grpcServer := grpc.NewServer()
	protos.RegisterShortenerServer(grpcServer, srv)

	log.Println("gRPC server is running")

	return grpcServer.Serve(lis)
}
