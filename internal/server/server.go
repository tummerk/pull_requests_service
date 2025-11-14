package server

import (
	"context"
	"errors"
	"fmt"
	"pull_requests_service/internal/domain"
	"pull_requests_service/internal/domain/entity"
	"pull_requests_service/internal/server/generated"
	"pull_requests_service/pkg/errcodes"
)

type PullRequestService interface {
	CreatePullRequest(ctx context.Context, pr entity.PullRequest) (entity.PullRequest, error)
	Merge(ctx context.Context, prId string) (entity.PullRequest, error)
	Reassign(ctx context.Context, prId string, oldReviewerId string) (entity.PullRequest, string, error)
}

// TeamService определяет бизнес-логику для работы с командами и их участниками.
type TeamService interface {
	TeamCreate(ctx context.Context, team entity.Team, users []entity.User) (entity.Team, []entity.User, error)
	TeamGet(ctx context.Context, name string) (entity.Team, []entity.User, error)
}

// UserService определяет бизнес-логику для работы с пользователями.
type UserService interface {
	CreateUser(ctx context.Context, user entity.User) (entity.User, error)
	GetByTeam(ctx context.Context, team string) ([]entity.User, error)
	SetIsActive(ctx context.Context, userId string, isActive bool) (entity.User, error)
}

type Server struct {
	prService   PullRequestService
	teamService TeamService
	userService UserService
}

func NewServer(prSvc PullRequestService, teamSvc TeamService, userSvc UserService) *Server {
	return &Server{
		prService:   prSvc,
		teamService: teamSvc,
		userService: userSvc,
	}
}

func (s *Server) PostPullRequestCreate(ctx context.Context, request generated.PostPullRequestCreateRequestObject) (generated.PostPullRequestCreateResponseObject, error) {
	prToCreate := entity.PullRequest{
		Id:       request.Body.PullRequestId,
		Name:     request.Body.PullRequestName,
		AuthorId: request.Body.AuthorId,
	}

	createdPR, err := s.prService.CreatePullRequest(ctx, prToCreate)
	if err != nil {

	}

	response := generated.PostPullRequestCreate201JSONResponse{
		Pr: &generated.PullRequest{
			PullRequestId:     createdPR.Id,
			PullRequestName:   createdPR.Name,
			AuthorId:          createdPR.AuthorId,
			AssignedReviewers: createdPR.AssignedReviewers,
			Status:            generated.PullRequestStatus(createdPR.Status),
			CreatedAt:         &createdPR.CreatedAt,
			MergedAt:          &createdPR.MergedAt,
		},
	}

	return response, nil
}

func (s *Server) GetTeamGet(ctx context.Context, request generated.GetTeamGetRequestObject) (generated.GetTeamGetResponseObject, error) {
	// 1. Параметры запроса уже в `request.Params`
	teamName := request.Params.TeamName

	// 2. Вызываем сервис
	team, users, err := s.teamService.TeamGet(ctx, teamName)
	if err != nil {
		// Обработка ошибки "не найдено"
		// if errors.Is(err, domain.ErrTeamNotFound) {
		//     return generated.GetTeamGet404JSONResponse{...}, nil
		// }
		return nil, err
	}

	// 3. Конвертируем результат в сгенерированные структуры и формируем ответ
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
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) PostPullRequestReassign(ctx context.Context, request generated.PostPullRequestReassignRequestObject) (generated.PostPullRequestReassignResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
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
			switch appErr.Code {
			case errcodes.TeamAlreadyExists, errcodes.UserAlreadyExists:
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

func (s *Server) GetUsersGetReview(ctx context.Context, request generated.GetUsersGetReviewRequestObject) (generated.GetUsersGetReviewResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) PostUsersSetIsActive(ctx context.Context, request generated.PostUsersSetIsActiveRequestObject) (generated.PostUsersSetIsActiveResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}
