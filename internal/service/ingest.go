package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"traffic-platform/internal/model"
	"traffic-platform/internal/store"
)

var (
	ErrInvalidEvent = errors.New("invalid traffic event")
)

// EventInput 是写入单个交通事件的入参结构（对齐 PRD 第 8 节请求示例）。
type EventInput struct {
	IntersectionID string    `json:"intersectionId"`
	Timestamp      time.Time `json:"timestamp"`
	VehicleCount   int       `json:"vehicleCount"`
	AvgSpeed       float64   `json:"avgSpeed"`
	Source         string    `json:"source"`
}

// SimulateInput 是生成模拟交通事件的入参。
type SimulateInput struct {
	Intersections   int `json:"intersections"`
	Minutes         int `json:"minutes"`
	EventsPerMinute int `json:"eventsPerMinute"`
}

// ImportError 记录单条导入失败的定位信息。
type ImportError struct {
	Row   int    `json:"row"`
	Error string `json:"error"`
}

// IngestService 负责事件接入与批量导入。
type IngestService struct {
	store store.Store
}

func NewIngestService(s store.Store) *IngestService {
	return &IngestService{store: s}
}

// IngestEvent 校验并写入单条交通事件。
func (s *IngestService) IngestEvent(ctx context.Context, in EventInput) (*model.RawTrafficEvent, error) {
	ev, err := buildEvent(in)
	if err != nil {
		return nil, err
	}
	if err := s.store.InsertEvent(ctx, ev); err != nil {
		return nil, err
	}
	return ev, nil
}

// ImportJSON 批量导入 JSON 数组并记录导入任务。
func (s *IngestService) ImportJSON(ctx context.Context, filename string, inputs []EventInput) (*model.ImportJob, error) {
	events := make([]*model.RawTrafficEvent, 0, len(inputs))
	var errs []ImportError
	buildFailed := 0
	for i, in := range inputs {
		ev, err := buildEvent(in)
		if err != nil {
			buildFailed++
			errs = append(errs, ImportError{Row: i + 1, Error: err.Error()})
			continue
		}
		events = append(events, ev)
	}
	success, err := s.store.InsertEvents(ctx, events)
	if err != nil {
		return nil, err
	}
	dbFailed := len(events) - success
	if dbFailed > 0 {
		errs = append(errs, ImportError{Row: 0, Error: fmt.Sprintf("database insert failed: %d row(s)", dbFailed)})
	}
	return s.finishImportJob(ctx, filename, len(inputs), success, buildFailed+dbFailed, errs)
}

// ImportCSV 解析 CSV 文件并导入，记录导入任务。
// CSV 列顺序：intersection_id,event_time,vehicle_count,avg_speed,source
func (s *IngestService) ImportCSV(ctx context.Context, filename string, r io.Reader) (*model.ImportJob, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}

	start := 0
	if len(records) > 0 && isHeaderRow(records[0]) {
		start = 1
	}

	events := make([]*model.RawTrafficEvent, 0, len(records))
	var errs []ImportError
	buildFailed := 0
	for i := start; i < len(records); i++ {
		rowNo := i + 1 // CSV 行号从 1 开始（含表头）
		in, err := parseCSVRow(records[i])
		if err != nil {
			buildFailed++
			errs = append(errs, ImportError{Row: rowNo, Error: err.Error()})
			continue
		}
		ev, err := buildEvent(in)
		if err != nil {
			buildFailed++
			errs = append(errs, ImportError{Row: rowNo, Error: err.Error()})
			continue
		}
		events = append(events, ev)
	}

	success, err := s.store.InsertEvents(ctx, events)
	if err != nil {
		return nil, err
	}
	dbFailed := len(events) - success
	if dbFailed > 0 {
		errs = append(errs, ImportError{Row: 0, Error: fmt.Sprintf("database insert failed: %d row(s)", dbFailed)})
	}

	total := len(records) - start
	if total < 0 {
		total = 0
	}
	return s.finishImportJob(ctx, filename, total, success, buildFailed+dbFailed, errs)
}

// ListImportJobs 返回导入任务列表。
func (s *IngestService) ListImportJobs(ctx context.Context) ([]*model.ImportJob, error) {
	return s.store.ListImportJobs(ctx)
}

// Simulate 生成一批模拟交通事件并写入数据库（演示/测试入口）。
func (s *IngestService) Simulate(ctx context.Context, in SimulateInput) (*model.ImportJob, error) {
	// 参数默认值与上限，避免生成过多数据
	if in.Intersections <= 0 {
		in.Intersections = 5
	}
	if in.Intersections > 50 {
		in.Intersections = 50
	}
	if in.Minutes <= 0 {
		in.Minutes = 10
	}
	if in.Minutes > 60 {
		in.Minutes = 60
	}
	if in.EventsPerMinute <= 0 {
		in.EventsPerMinute = 3
	}
	if in.EventsPerMinute > 20 {
		in.EventsPerMinute = 20
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	now := time.Now()
	events := make([]*model.RawTrafficEvent, 0, in.Intersections*in.Minutes*in.EventsPerMinute)
	for i := 0; i < in.Intersections; i++ {
		id := fmt.Sprintf("A-%d", 101+i)
		for m := in.Minutes - 1; m >= 0; m-- {
			for e := 0; e < in.EventsPerMinute; e++ {
				ts := now.Add(-time.Duration(m) * time.Minute).Add(-time.Duration(rng.Intn(60)) * time.Second)
				events = append(events, &model.RawTrafficEvent{
					IntersectionID: id,
					EventTime:      ts,
					VehicleCount:   20 + rng.Intn(60),
					AvgSpeed:       15 + rng.Float64()*45,
					Source:         "simulator",
				})
			}
		}
	}

	success, err := s.store.InsertEvents(ctx, events)
	if err != nil {
		return nil, err
	}
	failed := len(events) - success
	var errs []ImportError
	if failed > 0 {
		errs = append(errs, ImportError{Row: 0, Error: fmt.Sprintf("database insert failed: %d row(s)", failed)})
	}
	return s.finishImportJob(ctx, "simulator", len(events), success, failed, errs)
}

func (s *IngestService) finishImportJob(ctx context.Context, filename string, total, success, failed int, errs []ImportError) (*model.ImportJob, error) {
	status := "completed"
	if failed > 0 && success == 0 {
		status = "failed"
	} else if failed > 0 {
		status = "partial"
	}

	details := ""
	if len(errs) > 0 {
		if b, err := json.Marshal(errs); err == nil {
			details = string(b)
		}
	}

	job := &model.ImportJob{
		Filename:     filename,
		Status:       status,
		TotalRows:    total,
		SuccessRows:  success,
		FailedRows:   failed,
		ErrorDetails: details,
	}
	if err := s.store.InsertImportJob(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

// buildEvent 做基础校验并构造事件对象。
func buildEvent(in EventInput) (*model.RawTrafficEvent, error) {
	if strings.TrimSpace(in.IntersectionID) == "" {
		return nil, fmt.Errorf("%w: intersectionId is required", ErrInvalidEvent)
	}
	if in.VehicleCount < 0 {
		return nil, fmt.Errorf("%w: vehicleCount must be >= 0", ErrInvalidEvent)
	}
	if in.AvgSpeed < 0 {
		return nil, fmt.Errorf("%w: avgSpeed must be >= 0", ErrInvalidEvent)
	}
	ts := in.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}
	source := in.Source
	if source == "" {
		source = "unknown"
	}
	return &model.RawTrafficEvent{
		IntersectionID: in.IntersectionID,
		EventTime:      ts,
		VehicleCount:   in.VehicleCount,
		AvgSpeed:       in.AvgSpeed,
		Source:         source,
	}, nil
}

func isHeaderRow(row []string) bool {
	if len(row) == 0 {
		return false
	}
	h := strings.ToLower(strings.TrimSpace(row[0]))
	return h == "intersection_id" || h == "intersectionid" || h == "intersection_id\r"
}

func parseCSVRow(row []string) (EventInput, error) {
	if len(row) < 5 {
		return EventInput{}, errors.New("csv row must have 5 columns")
	}
	ts, err := time.Parse(time.RFC3339, strings.TrimSpace(row[1]))
	if err != nil {
		return EventInput{}, fmt.Errorf("invalid event_time: %w", err)
	}
	vc, err := strconv.Atoi(strings.TrimSpace(row[2]))
	if err != nil {
		return EventInput{}, fmt.Errorf("invalid vehicle_count: %w", err)
	}
	speed, err := strconv.ParseFloat(strings.TrimSpace(row[3]), 64)
	if err != nil {
		return EventInput{}, fmt.Errorf("invalid avg_speed: %w", err)
	}
	return EventInput{
		IntersectionID: strings.TrimSpace(row[0]),
		Timestamp:      ts,
		VehicleCount:   vc,
		AvgSpeed:       speed,
		Source:         strings.TrimSpace(row[4]),
	}, nil
}
