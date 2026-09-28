package service

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
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
	failed := 0
	for _, in := range inputs {
		ev, err := buildEvent(in)
		if err != nil {
			failed++
			continue
		}
		events = append(events, ev)
	}
	success, err := s.store.InsertEvents(ctx, events)
	if err != nil {
		return nil, err
	}
	failed += len(events) - success
	return s.finishImportJob(ctx, filename, len(inputs), success, failed)
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
	failed := 0
	for i := start; i < len(records); i++ {
		in, err := parseCSVRow(records[i])
		if err != nil {
			failed++
			continue
		}
		ev, err := buildEvent(in)
		if err != nil {
			failed++
			continue
		}
		events = append(events, ev)
	}

	success, err := s.store.InsertEvents(ctx, events)
	if err != nil {
		return nil, err
	}
	failed += len(events) - success

	total := len(records) - start
	if total < 0 {
		total = 0
	}
	return s.finishImportJob(ctx, filename, total, success, failed)
}

// ListImportJobs 返回导入任务列表。
func (s *IngestService) ListImportJobs(ctx context.Context) ([]*model.ImportJob, error) {
	return s.store.ListImportJobs(ctx)
}

func (s *IngestService) finishImportJob(ctx context.Context, filename string, total, success, failed int) (*model.ImportJob, error) {
	status := "completed"
	if failed > 0 && success == 0 {
		status = "failed"
	} else if failed > 0 {
		status = "partial"
	}
	job := &model.ImportJob{
		Filename:    filename,
		Status:      status,
		TotalRows:   total,
		SuccessRows: success,
		FailedRows:  failed,
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
