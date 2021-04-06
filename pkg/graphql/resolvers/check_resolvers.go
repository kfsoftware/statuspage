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
	"github.com/kfsoftware/statuspage/pkg/jobs"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type Resolver struct {
	Db       *gorm.DB
	Registry *jobs.SchedulerRegistry
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
	chk := db.Check{}
	resultDb := m.Db.First(&chk, "identifier = ?", input.ID)
	if resultDb.Error == nil {
		log.Infof("Check %s exists", input.ID)
		chk.Data = jsonBytes
		chk.Frecuency = input.Frecuency
		resultDb = m.Db.Save(chk)
		if resultDb.Error != nil {
			return nil, resultDb.Error
		}
		return mapCheck(chk)
	}
	checkId := uuid.New().String()
	result := m.Db.Create(&db.Check{
		ID:         checkId,
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		Data:       jsonBytes,
		Type:       check.TcpType,
		Status:     db.Scheduled,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	err = m.addCheckResult(checkId)
	if err != nil {
		return nil, err
	}
	return models.TCPCheck{
		ID:         input.ID,
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		Address:    input.Address,
	}, nil
}
func (m mutationResolver) addCheckResult(id string) error {
	chk := db.Check{}
	resultDb := m.Db.First(&chk, "id = ?", id)
	if resultDb.Error != nil {
		return resultDb.Error
	}
	interval, err := time.ParseDuration(chk.Frecuency)
	if err != nil {
		return err
	}
	chk.Check(m.Db)
	err = m.Registry.Register(chk.ID, interval, func() {
		chk.Check(m.Db)
	})
	if err != nil {
		return err
	}
	return nil
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
	chk := db.Check{}
	resultDb := m.Db.First(&chk, "identifier = ?", input.ID)
	if resultDb.Error == nil {
		log.Infof("Check %s exists", input.ID)
		chk.Data = jsonBytes
		chk.Frecuency = input.Frecuency
		resultDb = m.Db.Save(chk)
		if resultDb.Error != nil {
			return nil, resultDb.Error
		}
		return mapCheck(chk)
	}

	checkId := uuid.New().String()
	result := m.Db.Create(&db.Check{
		ID:         checkId,
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		Data:       jsonBytes,
		Type:       check.TlsType,
		Status:     db.Scheduled,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	err = m.addCheckResult(checkId)
	if err != nil {
		return nil, err
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
	chk := db.Check{}
	resultDb := m.Db.First(&chk, "identifier = ?", input.ID)
	if resultDb.Error == nil {
		log.Infof("Check %s exists", input.ID)
		chk.Data = jsonBytes
		chk.Frecuency = input.Frecuency
		resultDb = m.Db.Save(chk)
		if resultDb.Error != nil {
			return nil, resultDb.Error
		}
		return mapCheck(chk)
	}
	checkId := uuid.New().String()
	result := m.Db.Create(&db.Check{
		ID:         checkId,
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		Data:       jsonBytes,
		Type:       check.IcmpType,
		Status:     db.Scheduled,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	err = m.addCheckResult(checkId)
	if err != nil {
		return nil, err
	}
	return models.IcmpCheck{
		ID:         checkId,
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		Address:    input.Address,
	}, nil
}

func (m mutationResolver) DeleteCheck(ctx context.Context, id string) (*models.DeleteResponse, error) {
	chk := db.Check{}
	result := m.Db.First(&chk, "id = ?", id)
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
	err := m.Registry.Unregister(id)
	if err != nil {
		return nil, err
	}
	return &models.DeleteResponse{ID: id}, nil
}

func (m mutationResolver) CreateHTTPCheck(ctx context.Context, input models.CreateHTTPCheckInput) (models.Check, error) {
	data := db.HttpCheckData{Url: input.URL, StatusCode: input.StatusCode}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	_, err = time.ParseDuration(input.Frecuency)
	if err != nil {
		return nil, err
	}
	chk := db.Check{}
	resultDb := m.Db.First(&chk, "identifier = ?", input.ID)
	if resultDb.Error == nil {
		log.Infof("Check %s exists", input.ID)
		chk.Data = jsonBytes
		chk.Frecuency = input.Frecuency
		resultDb = m.Db.Save(chk)
		if resultDb.Error != nil {
			return nil, resultDb.Error
		}
		return mapCheck(chk)
	}

	checkId := uuid.New().String()
	result := m.Db.Create(&db.Check{
		ID:         checkId,
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		Data:       jsonBytes,
		Type:       check.HttpType,
		Status:     db.Scheduled,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	err = m.addCheckResult(checkId)
	if err != nil {
		return nil, err
	}
	return models.HTTPCheck{
		ID:         input.ID,
		Identifier: input.ID,
		Frecuency:  input.Frecuency,
		URL:        input.URL,
	}, nil
}

type queryResolver struct{ *Resolver }

func (q queryResolver) Check(ctx context.Context, checkID string) (models.Check, error) {
	chk := db.Check{}
	result := q.Db.First(&chk, "id = ?", checkID)
	if result.Error != nil {
		return nil, result.Error
	}
	return mapCheck(chk)
}

func (q queryResolver) Execution(ctx context.Context, execID string) (*models.CheckExecution, error) {
	chkExec := db.CheckExecution{}
	result := q.Db.First(&chkExec, "id = ?", execID)
	if result.Error != nil {
		return nil, result.Error
	}
	return mapCheckExecution(chkExec), nil
}

func (q queryResolver) Executions(ctx context.Context, checkID string, from *time.Time, until *time.Time) ([]*models.CheckExecution, error) {
	var executions []db.CheckExecution
	result := q.Db.Order("created_at desc").Where("check_id = ? AND created_at BETWEEN ? AND ?", checkID, from, until).Find(&executions)
	if result.Error != nil {
		return nil, result.Error
	}
	var modelExecutions []*models.CheckExecution
	for _, execution := range executions {
		modelExecutions = append(modelExecutions, mapCheckExecution(execution))
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
		modelChk, err := mapCheck(chk)
		if err != nil {
			return nil, err
		}
		modelChecks = append(modelChecks, modelChk)
	}
	return modelChecks, nil
}
func mapCheckExecution(chkExec db.CheckExecution) *models.CheckExecution {
	return &models.CheckExecution{
		ID:            chkExec.ID,
		ExecutionTime: chkExec.CreatedAt,
		Message:       chkExec.Message,
		ErrorMsg:      chkExec.ErrorMsg,
		Status:        string(chkExec.Status),
	}
}
func mapCheck(chk db.Check) (models.Check, error) {
	var modelCheck models.Check
	errorMsg := chk.ErrorMsg
	msg := chk.Message
	latestCheck := chk.LatestCheck

	switch chk.Type {
	case check.HttpType:
		httpCheckData, err := chk.GetHttpData()
		if err != nil {
			return nil, err
		}
		modelCheck = models.HTTPCheck{
			ID:          chk.ID,
			Identifier:  chk.Identifier,
			Frecuency:   chk.Frecuency,
			URL:         httpCheckData.Url,
			Status:      string(chk.Status),
			LatestCheck: &latestCheck,
			ErrorMsg:    errorMsg,
			Message:     msg,
		}
	case check.TcpType:
		tcpCheckData, err := chk.GetTcpData()
		if err != nil {
			return nil, err
		}
		modelCheck = models.TCPCheck{
			ID:          chk.ID,
			Identifier:  chk.Identifier,
			Frecuency:   chk.Frecuency,
			Address:     tcpCheckData.Address,
			Status:      string(chk.Status),
			LatestCheck: &latestCheck,
			ErrorMsg:    errorMsg,
			Message:     msg,
		}
	case check.TlsType:
		tlsCheckData, err := chk.GetTlsData()
		if err != nil {
			return nil, err
		}
		modelCheck = models.TCPCheck{
			ID:          chk.ID,
			Identifier:  chk.Identifier,
			Frecuency:   chk.Frecuency,
			Address:     tlsCheckData.Address,
			Status:      string(chk.Status),
			LatestCheck: &latestCheck,
			ErrorMsg:    errorMsg,
			Message:     msg,
		}
	case check.IcmpType:
		icmpCheckData, err := chk.GetIcmpData()
		if err != nil {
			return nil, err
		}
		modelCheck = models.TCPCheck{
			ID:          chk.ID,
			Identifier:  chk.Identifier,
			Frecuency:   chk.Frecuency,
			Address:     icmpCheckData.Address,
			Status:      string(chk.Status),
			LatestCheck: &latestCheck,
			ErrorMsg:    errorMsg,
			Message:     msg,
		}
	}
	return modelCheck, nil
}
