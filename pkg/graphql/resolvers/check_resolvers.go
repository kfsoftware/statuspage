package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/kfsoftware/statuspage/pkg/check"
	"github.com/kfsoftware/statuspage/pkg/db"
	"github.com/kfsoftware/statuspage/pkg/graphql/generated"
	"github.com/kfsoftware/statuspage/pkg/graphql/models"
	"gorm.io/gorm"
	"time"
)

type Resolver struct {
	Db *gorm.DB
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }

func (m mutationResolver) Poll(ctx context.Context) (*models.PollResult, error) {
	start := time.Now()
	db.CheckAll(m.Db)
	end := time.Now()
	return &models.PollResult{Took: int(end.Sub(start).Milliseconds())}, nil
}

func (m mutationResolver) CreateTCPCheck(ctx context.Context, input models.CreateTCPCheckInput) (models.Check, error) {
	data := db.TcpCheckData{Address: input.Address}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	_, err = time.ParseDuration(input.Frecuency)
	if err != nil {
		return nil, err
	}
	result := m.Db.Create(&db.Check{
		ID:         uuid.New().String(),
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		Data:       jsonBytes,
		Type:       check.TcpType,
		Status:     db.Scheduled,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	return models.TCPCheck{
		ID:         input.ID,
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		Address:    input.Address,
	}, nil
}

func (m mutationResolver) CreateTLSCheck(ctx context.Context, input models.CreateTLSCheckInput) (models.Check, error) {
	data := db.TlsCheckData{
		Address: input.Address,
		RootCAs: *input.RootCAs,
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	_, err = time.ParseDuration(input.Frecuency)
	if err != nil {
		return nil, err
	}
	result := m.Db.Create(&db.Check{
		ID:         uuid.New().String(),
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		Data:       jsonBytes,
		Type:       check.TlsType,
		Status:     db.Scheduled,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	return models.TLSCheck{
		ID:         input.ID,
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		Address:    input.Address,
	}, nil
}

func (m mutationResolver) CreateIcmpCheck(ctx context.Context, input models.CreateIcmpCheckInput) (models.Check, error) {
	data := db.IcmpCheckData{Address: input.Address}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	_, err = time.ParseDuration(input.Frecuency)
	if err != nil {
		return nil, err
	}
	result := m.Db.Create(&db.Check{
		ID:         uuid.New().String(),
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		Data:       jsonBytes,
		Type:       check.IcmpType,
		Status:     db.Scheduled,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	return models.IcmpCheck{
		ID:         input.ID,
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		Address:    input.Address,
	}, nil
}

func (m mutationResolver) DeleteCheck(ctx context.Context, id string) (*models.DeleteResponse, error) {
	chk := db.Check{}
	result := m.Db.Find(&chk, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	chk.Identifier = fmt.Sprintf("%s-%s-%s", chk.Identifier, "deleted", uuid.New().String())
	result = m.Db.Save(&chk)
	if result.Error != nil {
		return nil, result.Error
	}
	result = m.Db.Delete(&chk)
	if result.Error != nil {
		return nil, result.Error
	}
	return &models.DeleteResponse{ID: id}, nil
}

func (m mutationResolver) CreateHTTPCheck(ctx context.Context, input models.CreateHTTPCheckInput) (models.Check, error) {
	data := db.HttpCheckData{Url: input.URL}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	_, err = time.ParseDuration(input.Frecuency)
	if err != nil {
		return nil, err
	}
	result := m.Db.Create(&db.Check{
		ID:         uuid.New().String(),
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		Data:       jsonBytes,
		Type:       check.HttpType,
		Status:     db.Scheduled,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	return models.HTTPCheck{
		ID:         input.ID,
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		URL:        input.URL,
	}, nil

}

type queryResolver struct{ *Resolver }

func (q queryResolver) Executions(ctx context.Context, checkID string, from *time.Time, until *time.Time) ([]*models.CheckExecution, error) {
	var executions []db.CheckExecution
	result := q.Db.Where("check_id = ? AND created_at BETWEEN ? AND ?", checkID, from, until).Find(&executions)
	if result.Error != nil {
		return nil, result.Error
	}
	var modelExecutions []*models.CheckExecution
	for _, execution := range executions {
		modelExecutions = append(modelExecutions, &models.CheckExecution{
			ID:            execution.ID,
			ExecutionTime: execution.CreatedAt,
			Message:       execution.Message,
			ErrorMsg:      execution.ErrorMsg,
			Status:        string(execution.Status),
		})
	}
	return modelExecutions, nil

}

func (q queryResolver) Checks(ctx context.Context) ([]models.Check, error) {
	var checks []db.Check
	result := q.Db.Find(&checks)
	if result.Error != nil {
		return nil, result.Error
	}
	var modelChecks []models.Check
	for _, chk := range checks {
		errorMsg := chk.ErrorMsg
		msg := chk.Message
		latestCheck := chk.LatestCheck
		switch chk.Type {
		case check.HttpType:
			httpCheckData, err := chk.GetHttpData()
			if err != nil {
				return nil, err
			}
			modelChecks = append(modelChecks, models.HTTPCheck{
				ID:          chk.ID,
				Identifier:  chk.Identifier,
				Frecuency:   chk.Frecuency,
				URL:         httpCheckData.Url,
				Status:      string(chk.Status),
				LatestCheck: &latestCheck,
				ErrorMsg:    errorMsg,
				Message:     msg,
			})
		case check.TcpType:
			tcpCheckData, err := chk.GetTcpData()
			if err != nil {
				return nil, err
			}
			modelChecks = append(modelChecks, models.TCPCheck{
				ID:          chk.ID,
				Identifier:  chk.Identifier,
				Frecuency:   chk.Frecuency,
				Address:     tcpCheckData.Address,
				Status:      string(chk.Status),
				LatestCheck: &latestCheck,
				ErrorMsg:    errorMsg,
				Message:     msg,
			})
		case check.TlsType:
			tlsCheckData, err := chk.GetTlsData()
			if err != nil {
				return nil, err
			}
			modelChecks = append(modelChecks, models.TCPCheck{
				ID:          chk.ID,
				Identifier:  chk.Identifier,
				Frecuency:   chk.Frecuency,
				Address:     tlsCheckData.Address,
				Status:      string(chk.Status),
				LatestCheck: &latestCheck,
				ErrorMsg:    errorMsg,
				Message:     msg,
			})
		case check.IcmpType:
			icmpCheckData, err := chk.GetIcmpData()
			if err != nil {
				return nil, err
			}
			modelChecks = append(modelChecks, models.TCPCheck{
				ID:          chk.ID,
				Identifier:  chk.Identifier,
				Frecuency:   chk.Frecuency,
				Address:     icmpCheckData.Address,
				Status:      string(chk.Status),
				LatestCheck: &latestCheck,
				ErrorMsg:    errorMsg,
				Message:     msg,
			})
		}

	}
	return modelChecks, nil
}
