package server

import (
	"context"
	"errors"
	"pull_requests_service/internal/domain"
	"pull_requests_service/internal/domain/entity"
	"pull_requests_service/internal/server/generated"
	"pull_requests_service/pkg/errcodes"
)

type PullRequestService interface {
	CreatePullRequest(ctx context.Context, pr entity.PullRequest) (entity.PullRequest, error)
	Merge(ctx context.Context, prId string) (entity.PullRequest, error)
	Reassign(ctx context.Context, prId string, oldReviewerId string) (entity.PullRequest, string, error)
	GetUserReviews(ctx context.Context, userId string) ([]entity.PullRequest, error)
}

// TeamService определяет бизнес-логику для работы с командами и их участниками.
type TeamService interface {
	TeamCreate(ctx context.Context, team entity.Team, users []entity.User) (entity.Team, []entity.User, error)
	TeamGet(ctx context.Context, name string) (entity.Team, []entity.User, error)
}

// UserService определяет бизнес-логику для работы с пользователями.
type UserService interface {
	SetIsActive(ctx context.Context, userId string, isActive bool) (entity.User, error)
}

type StatsService interface {
	GetUserAssignmentStats(ctx context.Context) ([]entity.UserAssignmentStat, error)
}

type Server struct {
	prService    PullRequestService
	teamService  TeamService
	userService  UserService
	statsService StatsService
}

func NewServer(prSvc PullRequestService, teamSvc TeamService, userSvc UserService, statSvc StatsService) *Server {
	return &Server{
		prService:    prSvc,
		teamService:  teamSvc,
		userService:  userSvc,
		statsService: statSvc,
	}
}

func (s *Server) PostPullRequestCreate(ctx context.Context,
	request generated.PostPullRequestCreateRequestObject) (generated.PostPullRequestCreateResponseObject, error) {

	prToCreate := entity.PullRequest{
		Id:       request.Body.PullRequestId,
		Name:     request.Body.PullRequestName,
		AuthorId: request.Body.AuthorId,
	}

	createdPR, err := s.prService.CreatePullRequest(ctx, prToCreate)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			switch appErr.Code {
			case errcodes.NotFound:
				return generated.PostPullRequestCreate404JSONResponse{
					Error: struct {
						Code    generated.ErrorResponseErrorCode `json:"code"`
						Message string                           `json:"message"`
					}{Code: generated.NOTFOUND, Message: appErr.Message},
				}, nil
			case errcodes.PullRequestExists:
				return generated.PostPullRequestCreate409JSONResponse{
					Error: struct {
						Code    generated.ErrorResponseErrorCode `json:"code"`
						Message string                           `json:"message"`
					}{Code: generated.PREXISTS, Message: appErr.Message},
				}, nil
			}
		}
		return nil, err
	}
	response := generated.PostPullRequestCreate201JSONResponse{
		Pr: &generated.PullRequest{
			PullRequestId:     createdPR.Id,
			PullRequestName:   createdPR.Name,
			AuthorId:          createdPR.AuthorId,
			AssignedReviewers: createdPR.AssignedReviewers,
			Status:            generated.PullRequestStatus(createdPR.Status),
		},
	}

	return response, nil
}

func (s *Server) GetTeamGet(ctx context.Context, request generated.GetTeamGetRequestObject) (generated.GetTeamGetResponseObject, error) {
	teamName := request.Params.TeamName
	team, users, err := s.teamService.TeamGet(ctx, teamName)
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		var response generated.GetTeamGet404JSONResponse
		response = generated.GetTeamGet404JSONResponse{
			Error: struct {
				Code    generated.ErrorResponseErrorCode `json:"code"`
				Message string                           `json:"message"`
			}{
				Code:    generated.ErrorResponseErrorCode(appErr.Code),
				Message: appErr.Error(),
			}}
		return response, nil
	}

	members := make([]generated.TeamMember, 0, len(users))
	for _, u := range users {
		members = append(members, generated.TeamMember{
			UserId:   u.Id,
			Username: u.Name,
			IsActive: u.IsActive,
		})
	}

	response := generated.GetTeamGet200JSONResponse{
		TeamName: team.Name,
		Members:  members,
	}

	return response, nil
}

func (s *Server) PostPullRequestMerge(ctx context.Context, request generated.PostPullRequestMergeRequestObject) (generated.PostPullRequestMergeResponseObject, error) {
	prId := request.Body.PullRequestId
	pr, err := s.prService.Merge(ctx, prId)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			switch appErr.Code {
			case errcodes.NotFound:
				response := generated.PostPullRequestMerge404JSONResponse{
					Error: struct {
						Code    generated.ErrorResponseErrorCode `json:"code"`
						Message string                           `json:"message"`
					}{
						Code:    generated.ErrorResponseErrorCode(appErr.Code),
						Message: appErr.Error(),
					}}
				return response, nil
			}
		}
	}
	response := generated.PostPullRequestMerge200JSONResponse{
		Pr: &generated.PullRequest{
			AssignedReviewers: pr.AssignedReviewers,
			AuthorId:          pr.AuthorId,
			CreatedAt:         &pr.CreatedAt,
			MergedAt:          pr.MergedAt,
			PullRequestId:     pr.Id,
			PullRequestName:   pr.Name,
			Status:            generated.PullRequestStatus(pr.Status),
		},
	}
	return response, nil
}

func (s *Server) PostPullRequestReassign(ctx context.Context, request generated.PostPullRequestReassignRequestObject) (generated.PostPullRequestReassignResponseObject, error) {
	prId := request.Body.PullRequestId
	oldUserId := request.Body.OldUserId
	pr, newId, err := s.prService.Reassign(ctx, prId, oldUserId)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			switch appErr.Code {
			case errcodes.NotFound:
				return generated.PostPullRequestReassign404JSONResponse{
					Error: struct {
						Code    generated.ErrorResponseErrorCode `json:"code"`
						Message string                           `json:"message"`
					}{Code: generated.NOTFOUND, Message: appErr.Message},
				}, nil

			case errcodes.PrMerged:
				return generated.PostPullRequestReassign409JSONResponse{
					Error: struct {
						Code    generated.ErrorResponseErrorCode `json:"code"`
						Message string                           `json:"message"`
					}{Code: generated.PRMERGED, Message: appErr.Message},
				}, nil
			case errcodes.NotAssigned:
				return generated.PostPullRequestReassign409JSONResponse{
					Error: struct {
						Code    generated.ErrorResponseErrorCode `json:"code"`
						Message string                           `json:"message"`
					}{Code: generated.NOTASSIGNED, Message: appErr.Message},
				}, nil
			case errcodes.NoCandidate:
				return generated.PostPullRequestReassign409JSONResponse{
					Error: struct {
						Code    generated.ErrorResponseErrorCode `json:"code"`
						Message string                           `json:"message"`
					}{Code: generated.NOCANDIDATE, Message: appErr.Message},
				}, nil
			}
		}
		return nil, err
	}
	response := generated.PostPullRequestReassign200JSONResponse{
		Pr: generated.PullRequest{
			AssignedReviewers: pr.AssignedReviewers,
			AuthorId:          pr.AuthorId,
			CreatedAt:         &pr.CreatedAt,
			MergedAt:          pr.MergedAt,
			PullRequestId:     pr.Id,
			PullRequestName:   pr.Name,
			Status:            generated.PullRequestStatus(pr.Status),
		},
		ReplacedBy: newId,
	}
	return response, nil
}

func (s *Server) PostTeamAdd(ctx context.Context, request generated.PostTeamAddRequestObject) (generated.PostTeamAddResponseObject, error) {
	if request.Body == nil {
		response := generated.PostTeamAdd400JSONResponse{
			Error: struct {
				Code    generated.ErrorResponseErrorCode `json:"code"`
				Message string                           `json:"message"`
			}{
				Code:    "INVALID_ARGUMENT",
				Message: "request body cannot be empty",
			},
		}
		return response, nil
	}

	domainTeam := entity.Team{
		Name: request.Body.TeamName,
	}

	domainUsers := make([]entity.User, len(request.Body.Members))
	for i, member := range request.Body.Members {
		domainUsers[i] = entity.User{
			Id:       member.UserId,
			Name:     member.Username,
			IsActive: member.IsActive,
			Team:     request.Body.TeamName,
		}
	}

	createdTeam, createdUsers, err := s.teamService.TeamCreate(ctx, domainTeam, domainUsers)

	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			response := generated.PostTeamAdd400JSONResponse{
				Error: struct {
					Code    generated.ErrorResponseErrorCode `json:"code"`
					Message string                           `json:"message"`
				}{
					Code:    generated.ErrorResponseErrorCode(appErr.Code),
					Message: appErr.Message,
				},
			}
			return response, nil

		}
		return nil, err
	}

	apiMembers := make([]generated.TeamMember, len(createdUsers))
	for i, u := range createdUsers {
		apiMembers[i] = generated.TeamMember{
			UserId:   u.Id,
			Username: u.Name,
			IsActive: u.IsActive,
		}
	}

	response := generated.PostTeamAdd201JSONResponse{
		Team: &generated.Team{
			TeamName: createdTeam.Name,
			Members:  apiMembers,
		},
	}
	return response, nil
}

func (s *Server) PostUsersSetIsActive(ctx context.Context, request generated.PostUsersSetIsActiveRequestObject) (generated.PostUsersSetIsActiveResponseObject, error) {
	isActive := request.Body.IsActive
	userId := request.Body.UserId
	user, err := s.userService.SetIsActive(ctx, userId, isActive)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			response := generated.PostUsersSetIsActive404JSONResponse{
				Error: struct {
					Code    generated.ErrorResponseErrorCode `json:"code"`
					Message string                           `json:"message"`
				}{
					Code:    generated.ErrorResponseErrorCode(appErr.Code),
					Message: appErr.Message,
				},
			}
			return response, nil
		}
		return nil, err
	}
	response := generated.PostUsersSetIsActive200JSONResponse{
		User: &generated.User{
			UserId:   user.Id,
			Username: user.Name,
			TeamName: user.Team,
			IsActive: user.IsActive,
		},
	}
	return response, nil
}

func (s *Server) GetUsersGetReview(ctx context.Context, request generated.GetUsersGetReviewRequestObject) (generated.GetUsersGetReviewResponseObject, error) {
	userId := request.Params.UserId
	prs, err := s.prService.GetUserReviews(ctx, userId)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			return nil, err
		}
	}
	var response generated.GetUsersGetReview200JSONResponse
	response.UserId = userId
	for _, pr := range prs {
		response.PullRequests = append(response.PullRequests, generated.PullRequestShort{
			AuthorId:        pr.AuthorId,
			PullRequestId:   pr.Id,
			PullRequestName: pr.Name,
			Status:          generated.PullRequestShortStatus(pr.Status),
		})
	}
	return response, nil
}

// статистика по кол-ву назначений пользователей
func (s *Server) GetUserStats(ctx context.Context, request generated.GetUserStatsRequestObject) (
	generated.GetUserStatsResponseObject, error) {

	stats, err := s.statsService.GetUserAssignmentStats(ctx)
	if err != nil {
		return nil, err
	}

	apiStats := make([]generated.UserAssignmentStat, 0, len(stats))

	for _, stat := range stats {
		apiStats = append(apiStats, generated.UserAssignmentStat{
			AssignmentCount: int32(stat.AssignmentCount),
			UserId:          stat.UserID,
			Username:        stat.Username,
		})
	}

	response := generated.GetUserStats200JSONResponse{
		AssignmentsByUser: &apiStats,
	}

	return response, nil
}
