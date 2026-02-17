package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gclassroom "google.golang.org/api/classroom/v1"

	"github.com/timothy/gc-cli/internal/classroom"
	"github.com/timothy/gc-cli/internal/store"
)

type Engine struct {
	Classroom classroom.ClassroomClient
	Store     *store.SQLiteStore
	Now       func() time.Time
}

func NewEngine(client classroom.ClassroomClient, st *store.SQLiteStore) *Engine {
	return &Engine{
		Classroom: client,
		Store:     st,
		Now:       func() time.Time { return time.Now().UTC() },
	}
}

func (e *Engine) RefreshCourses(ctx context.Context) ([]classroom.CourseSnapshot, error) {
	courses, _, err := e.Classroom.ListCourses(ctx, classroom.ListParams{PageSize: 200})
	if err != nil {
		return nil, err
	}
	snapshots := make([]classroom.CourseSnapshot, 0, len(courses))
	for _, c := range courses {
		snapshot := classroom.CourseSnapshot{
			ID:             c.Id,
			Name:           c.Name,
			Section:        c.Section,
			Description:    c.Description,
			OwnerID:        c.OwnerId,
			EnrollmentCode: c.EnrollmentCode,
			State:          c.CourseState,
			UpdatedAt:      parseTimestamp(c.UpdateTime),
		}
		snapshots = append(snapshots, snapshot)
		payload, _ := json.Marshal(snapshot)
		if err := e.Store.UpsertCacheEntry(ctx, store.CacheEntry{
			Kind:      "course",
			Key:       c.Id,
			Data:      payload,
			ETag:      "",
			UpdatedAt: e.Now(),
			Stale:     false,
		}); err != nil {
			return nil, err
		}
	}
	if err := e.Store.SetLastRefresh(ctx, "courses", e.Now()); err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (e *Engine) RefreshCourseWork(ctx context.Context, courseID string) ([]classroom.CourseWorkSnapshot, error) {
	items, _, err := e.Classroom.ListCourseWork(ctx, courseID, classroom.ListParams{PageSize: 200})
	if err != nil {
		return nil, err
	}
	out := make([]classroom.CourseWorkSnapshot, 0, len(items))
	for _, item := range items {
		snapshot := classroom.CourseWorkSnapshot{
			CourseID:     courseID,
			CourseWorkID: item.Id,
			Title:        item.Title,
			Description:  item.Description,
			WorkType:     item.WorkType,
			State:        item.State,
			TopicID:      item.TopicId,
			MaxPoints:    item.MaxPoints,
			DueAt:        parseDueDate(item),
			UpdatedAt:    parseTimestamp(item.UpdateTime),
		}
		out = append(out, snapshot)
		payload, _ := json.Marshal(snapshot)
		if err := e.Store.UpsertCacheEntry(ctx, store.CacheEntry{
			Kind:      "coursework",
			Key:       fmt.Sprintf("%s:%s", courseID, item.Id),
			Data:      payload,
			ETag:      "",
			UpdatedAt: e.Now(),
			Stale:     false,
		}); err != nil {
			return nil, err
		}
	}
	if err := e.Store.SetLastRefresh(ctx, fmt.Sprintf("coursework:%s", courseID), e.Now()); err != nil {
		return nil, err
	}
	return out, nil
}

func (e *Engine) RefreshSubmissions(ctx context.Context, courseID, courseWorkID string) ([]classroom.StudentSubmissionSnapshot, error) {
	subs, _, err := e.Classroom.ListStudentSubmissions(ctx, courseID, courseWorkID, "", classroom.ListParams{PageSize: 200})
	if err != nil {
		return nil, err
	}
	out := make([]classroom.StudentSubmissionSnapshot, 0, len(subs))
	for _, sub := range subs {
		snapshot := classroom.StudentSubmissionSnapshot{
			CourseID:      courseID,
			CourseWorkID:  courseWorkID,
			SubmissionID:  sub.Id,
			UserID:        sub.UserId,
			State:         sub.State,
			AssignedGrade: sub.AssignedGrade,
			DraftGrade:    sub.DraftGrade,
			Late:          sub.Late,
			UpdateTime:    parseTimestamp(sub.UpdateTime),
		}
		out = append(out, snapshot)
		payload, _ := json.Marshal(snapshot)
		if err := e.Store.UpsertCacheEntry(ctx, store.CacheEntry{
			Kind:      "submission",
			Key:       fmt.Sprintf("%s:%s:%s", courseID, courseWorkID, sub.Id),
			Data:      payload,
			ETag:      "",
			UpdatedAt: e.Now(),
			Stale:     false,
		}); err != nil {
			return nil, err
		}
	}
	if err := e.Store.SetLastRefresh(ctx, fmt.Sprintf("submissions:%s:%s", courseID, courseWorkID), e.Now()); err != nil {
		return nil, err
	}
	return out, nil
}

func (e *Engine) MarkKindStale(ctx context.Context, kind string) error {
	return e.Store.MarkStale(ctx, kind, "")
}

func parseTimestamp(raw string) time.Time {
	if raw == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseDueDate(work *gclassroom.CourseWork) time.Time {
	if work == nil || work.DueDate == nil {
		return time.Time{}
	}
	hour, minute, second := 23, 59, 59
	if work.DueTime != nil {
		hour = int(work.DueTime.Hours)
		minute = int(work.DueTime.Minutes)
		second = int(work.DueTime.Seconds)
	}
	return time.Date(int(work.DueDate.Year), time.Month(work.DueDate.Month), int(work.DueDate.Day), hour, minute, second, 0, time.Local)
}
