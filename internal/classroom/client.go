package classroom

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"golang.org/x/oauth2"
	gclassroom "google.golang.org/api/classroom/v1"
	"google.golang.org/api/option"
)

type ClassroomClient interface {
	ListCourses(ctx context.Context, params ListParams) ([]*gclassroom.Course, string, error)
	GetCourse(ctx context.Context, courseID string) (*gclassroom.Course, error)
	CreateCourse(ctx context.Context, course *gclassroom.Course) (*gclassroom.Course, error)
	UpdateCourse(ctx context.Context, courseID string, patch *gclassroom.Course, updateMask []string) (*gclassroom.Course, error)
	SetCourseState(ctx context.Context, courseID, state string) (*gclassroom.Course, error)
	DeleteCourse(ctx context.Context, courseID string) error

	ListAnnouncements(ctx context.Context, courseID string, params ListParams) ([]*gclassroom.Announcement, string, error)
	CreateAnnouncement(ctx context.Context, courseID string, announcement *gclassroom.Announcement) (*gclassroom.Announcement, error)
	UpdateAnnouncement(ctx context.Context, courseID, announcementID string, patch *gclassroom.Announcement, updateMask []string) (*gclassroom.Announcement, error)
	DeleteAnnouncement(ctx context.Context, courseID, announcementID string) error

	ListCourseWork(ctx context.Context, courseID string, params ListParams) ([]*gclassroom.CourseWork, string, error)
	GetCourseWork(ctx context.Context, courseID, courseWorkID string) (*gclassroom.CourseWork, error)
	CreateCourseWork(ctx context.Context, courseID string, work *gclassroom.CourseWork) (*gclassroom.CourseWork, error)
	PatchCourseWork(ctx context.Context, courseID, courseWorkID string, patch *gclassroom.CourseWork, updateMask []string) (*gclassroom.CourseWork, error)
	PublishCourseWork(ctx context.Context, courseID, courseWorkID string) (*gclassroom.CourseWork, error)
	ScheduleCourseWork(ctx context.Context, courseID, courseWorkID string, when time.Time) (*gclassroom.CourseWork, error)
	DeleteCourseWork(ctx context.Context, courseID, courseWorkID string) error

	ListStudentSubmissions(ctx context.Context, courseID, courseWorkID, userID string, params ListParams) ([]*gclassroom.StudentSubmission, string, error)
	GetStudentSubmission(ctx context.Context, courseID, courseWorkID, submissionID string) (*gclassroom.StudentSubmission, error)
	TurnInSubmission(ctx context.Context, courseID, courseWorkID, submissionID string) (*gclassroom.StudentSubmission, error)
	ReclaimSubmission(ctx context.Context, courseID, courseWorkID, submissionID string) (*gclassroom.StudentSubmission, error)
	ReturnSubmission(ctx context.Context, courseID, courseWorkID, submissionID string) (*gclassroom.StudentSubmission, error)
	PatchStudentSubmissionGrades(ctx context.Context, courseID, courseWorkID, submissionID string, draftGrade, assignedGrade *float64) (*gclassroom.StudentSubmission, error)

	ListTeachers(ctx context.Context, courseID string, params ListParams) ([]*gclassroom.Teacher, string, error)
	ListStudents(ctx context.Context, courseID string, params ListParams) ([]*gclassroom.Student, string, error)
	InviteTeacher(ctx context.Context, courseID, userID string) (*gclassroom.Invitation, error)
	InviteStudent(ctx context.Context, courseID, userID string) (*gclassroom.Invitation, error)
	RemoveTeacher(ctx context.Context, courseID, userID string) error
	RemoveStudent(ctx context.Context, courseID, userID string) error

	ListTopics(ctx context.Context, courseID string, params ListParams) ([]*gclassroom.Topic, string, error)
	CreateTopic(ctx context.Context, courseID, name string) (*gclassroom.Topic, error)
	UpdateTopic(ctx context.Context, courseID, topicID, name string) (*gclassroom.Topic, error)
	DeleteTopic(ctx context.Context, courseID, topicID string) error
	MoveCourseWorkToTopic(ctx context.Context, courseID, courseWorkID, topicID string) (*gclassroom.CourseWork, error)

	BuildTodo(ctx context.Context) ([]TodoItem, error)
}

type Client struct {
	svc *gclassroom.Service
}

func NewClient(ctx context.Context, tokenSource oauth2.TokenSource) (*Client, error) {
	httpClient := oauth2.NewClient(ctx, tokenSource)
	httpClient.Transport = newRetryTransport(httpClient.Transport)

	svc, err := gclassroom.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create classroom service: %w", err)
	}
	return &Client{svc: svc}, nil
}

func (c *Client) ListCourses(ctx context.Context, params ListParams) ([]*gclassroom.Course, string, error) {
	call := c.svc.Courses.List().Context(ctx)
	if params.PageSize > 0 {
		call = call.PageSize(params.PageSize)
	}
	if params.PageToken != "" {
		call = call.PageToken(params.PageToken)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("list courses: %w", err)
	}
	return resp.Courses, resp.NextPageToken, nil
}

func (c *Client) GetCourse(ctx context.Context, courseID string) (*gclassroom.Course, error) {
	course, err := c.svc.Courses.Get(courseID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get course %s: %w", courseID, err)
	}
	return course, nil
}

func (c *Client) CreateCourse(ctx context.Context, course *gclassroom.Course) (*gclassroom.Course, error) {
	out, err := c.svc.Courses.Create(course).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create course: %w", err)
	}
	return out, nil
}

func (c *Client) UpdateCourse(ctx context.Context, courseID string, patch *gclassroom.Course, updateMask []string) (*gclassroom.Course, error) {
	call := c.svc.Courses.Patch(courseID, patch).Context(ctx)
	if len(updateMask) > 0 {
		call = call.UpdateMask(strings.Join(updateMask, ","))
	}
	out, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("update course %s: %w", courseID, err)
	}
	return out, nil
}

func (c *Client) SetCourseState(ctx context.Context, courseID, state string) (*gclassroom.Course, error) {
	patch := &gclassroom.Course{CourseState: state}
	out, err := c.svc.Courses.Patch(courseID, patch).UpdateMask("courseState").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("set course state %s for %s: %w", state, courseID, err)
	}
	return out, nil
}

func (c *Client) DeleteCourse(ctx context.Context, courseID string) error {
	if _, err := c.svc.Courses.Delete(courseID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete course %s: %w", courseID, err)
	}
	return nil
}

func (c *Client) ListAnnouncements(ctx context.Context, courseID string, params ListParams) ([]*gclassroom.Announcement, string, error) {
	call := c.svc.Courses.Announcements.List(courseID).Context(ctx)
	if params.PageSize > 0 {
		call = call.PageSize(params.PageSize)
	}
	if params.PageToken != "" {
		call = call.PageToken(params.PageToken)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("list announcements: %w", err)
	}
	return resp.Announcements, resp.NextPageToken, nil
}

func (c *Client) CreateAnnouncement(ctx context.Context, courseID string, announcement *gclassroom.Announcement) (*gclassroom.Announcement, error) {
	out, err := c.svc.Courses.Announcements.Create(courseID, announcement).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create announcement: %w", err)
	}
	return out, nil
}

func (c *Client) UpdateAnnouncement(ctx context.Context, courseID, announcementID string, patch *gclassroom.Announcement, updateMask []string) (*gclassroom.Announcement, error) {
	call := c.svc.Courses.Announcements.Patch(courseID, announcementID, patch).Context(ctx)
	if len(updateMask) > 0 {
		call = call.UpdateMask(strings.Join(updateMask, ","))
	}
	out, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("update announcement %s: %w", announcementID, err)
	}
	return out, nil
}

func (c *Client) DeleteAnnouncement(ctx context.Context, courseID, announcementID string) error {
	if _, err := c.svc.Courses.Announcements.Delete(courseID, announcementID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete announcement %s: %w", announcementID, err)
	}
	return nil
}

func (c *Client) ListCourseWork(ctx context.Context, courseID string, params ListParams) ([]*gclassroom.CourseWork, string, error) {
	call := c.svc.Courses.CourseWork.List(courseID).Context(ctx)
	if params.PageSize > 0 {
		call = call.PageSize(params.PageSize)
	}
	if params.PageToken != "" {
		call = call.PageToken(params.PageToken)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("list coursework: %w", err)
	}
	return resp.CourseWork, resp.NextPageToken, nil
}

func (c *Client) GetCourseWork(ctx context.Context, courseID, courseWorkID string) (*gclassroom.CourseWork, error) {
	out, err := c.svc.Courses.CourseWork.Get(courseID, courseWorkID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get coursework %s: %w", courseWorkID, err)
	}
	return out, nil
}

func (c *Client) CreateCourseWork(ctx context.Context, courseID string, work *gclassroom.CourseWork) (*gclassroom.CourseWork, error) {
	out, err := c.svc.Courses.CourseWork.Create(courseID, work).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create coursework: %w", err)
	}
	return out, nil
}

func (c *Client) PatchCourseWork(ctx context.Context, courseID, courseWorkID string, patch *gclassroom.CourseWork, updateMask []string) (*gclassroom.CourseWork, error) {
	call := c.svc.Courses.CourseWork.Patch(courseID, courseWorkID, patch).Context(ctx)
	if len(updateMask) > 0 {
		call = call.UpdateMask(strings.Join(updateMask, ","))
	}
	out, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("patch coursework %s: %w", courseWorkID, err)
	}
	return out, nil
}

func (c *Client) PublishCourseWork(ctx context.Context, courseID, courseWorkID string) (*gclassroom.CourseWork, error) {
	patch := &gclassroom.CourseWork{State: "PUBLISHED"}
	return c.PatchCourseWork(ctx, courseID, courseWorkID, patch, []string{"state"})
}

func (c *Client) ScheduleCourseWork(ctx context.Context, courseID, courseWorkID string, when time.Time) (*gclassroom.CourseWork, error) {
	patch := &gclassroom.CourseWork{State: "PUBLISHED", ScheduledTime: when.Format(time.RFC3339)}
	return c.PatchCourseWork(ctx, courseID, courseWorkID, patch, []string{"state", "scheduledTime"})
}

func (c *Client) DeleteCourseWork(ctx context.Context, courseID, courseWorkID string) error {
	if _, err := c.svc.Courses.CourseWork.Delete(courseID, courseWorkID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete coursework %s: %w", courseWorkID, err)
	}
	return nil
}

func (c *Client) ListStudentSubmissions(ctx context.Context, courseID, courseWorkID, userID string, params ListParams) ([]*gclassroom.StudentSubmission, string, error) {
	call := c.svc.Courses.CourseWork.StudentSubmissions.List(courseID, courseWorkID).Context(ctx)
	if userID != "" {
		call = call.UserId(userID)
	}
	if params.PageSize > 0 {
		call = call.PageSize(params.PageSize)
	}
	if params.PageToken != "" {
		call = call.PageToken(params.PageToken)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("list submissions: %w", err)
	}
	return resp.StudentSubmissions, resp.NextPageToken, nil
}

func (c *Client) GetStudentSubmission(ctx context.Context, courseID, courseWorkID, submissionID string) (*gclassroom.StudentSubmission, error) {
	out, err := c.svc.Courses.CourseWork.StudentSubmissions.Get(courseID, courseWorkID, submissionID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get submission %s: %w", submissionID, err)
	}
	return out, nil
}

func (c *Client) TurnInSubmission(ctx context.Context, courseID, courseWorkID, submissionID string) (*gclassroom.StudentSubmission, error) {
	_, err := c.svc.Courses.CourseWork.StudentSubmissions.TurnIn(courseID, courseWorkID, submissionID, &gclassroom.TurnInStudentSubmissionRequest{}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("turn in submission %s: %w", submissionID, err)
	}
	return c.GetStudentSubmission(ctx, courseID, courseWorkID, submissionID)
}

func (c *Client) ReclaimSubmission(ctx context.Context, courseID, courseWorkID, submissionID string) (*gclassroom.StudentSubmission, error) {
	_, err := c.svc.Courses.CourseWork.StudentSubmissions.Reclaim(courseID, courseWorkID, submissionID, &gclassroom.ReclaimStudentSubmissionRequest{}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("reclaim submission %s: %w", submissionID, err)
	}
	return c.GetStudentSubmission(ctx, courseID, courseWorkID, submissionID)
}

func (c *Client) ReturnSubmission(ctx context.Context, courseID, courseWorkID, submissionID string) (*gclassroom.StudentSubmission, error) {
	_, err := c.svc.Courses.CourseWork.StudentSubmissions.Return(courseID, courseWorkID, submissionID, &gclassroom.ReturnStudentSubmissionRequest{}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("return submission %s: %w", submissionID, err)
	}
	return c.GetStudentSubmission(ctx, courseID, courseWorkID, submissionID)
}

func (c *Client) PatchStudentSubmissionGrades(ctx context.Context, courseID, courseWorkID, submissionID string, draftGrade, assignedGrade *float64) (*gclassroom.StudentSubmission, error) {
	patch := &gclassroom.StudentSubmission{}
	masks := make([]string, 0, 2)
	if draftGrade != nil {
		patch.DraftGrade = *draftGrade
		masks = append(masks, "draftGrade")
	}
	if assignedGrade != nil {
		patch.AssignedGrade = *assignedGrade
		masks = append(masks, "assignedGrade")
	}
	if len(masks) == 0 {
		return c.GetStudentSubmission(ctx, courseID, courseWorkID, submissionID)
	}
	_, err := c.svc.Courses.CourseWork.StudentSubmissions.Patch(courseID, courseWorkID, submissionID, patch).UpdateMask(strings.Join(masks, ",")).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("grade submission %s: %w", submissionID, err)
	}
	return c.GetStudentSubmission(ctx, courseID, courseWorkID, submissionID)
}

func (c *Client) ListTeachers(ctx context.Context, courseID string, params ListParams) ([]*gclassroom.Teacher, string, error) {
	call := c.svc.Courses.Teachers.List(courseID).Context(ctx)
	if params.PageSize > 0 {
		call = call.PageSize(params.PageSize)
	}
	if params.PageToken != "" {
		call = call.PageToken(params.PageToken)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("list teachers: %w", err)
	}
	return resp.Teachers, resp.NextPageToken, nil
}

func (c *Client) ListStudents(ctx context.Context, courseID string, params ListParams) ([]*gclassroom.Student, string, error) {
	call := c.svc.Courses.Students.List(courseID).Context(ctx)
	if params.PageSize > 0 {
		call = call.PageSize(params.PageSize)
	}
	if params.PageToken != "" {
		call = call.PageToken(params.PageToken)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("list students: %w", err)
	}
	return resp.Students, resp.NextPageToken, nil
}

func (c *Client) InviteTeacher(ctx context.Context, courseID, userID string) (*gclassroom.Invitation, error) {
	inv, err := c.svc.Invitations.Create(&gclassroom.Invitation{CourseId: courseID, Role: "TEACHER", UserId: userID}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("invite teacher: %w", err)
	}
	return inv, nil
}

func (c *Client) InviteStudent(ctx context.Context, courseID, userID string) (*gclassroom.Invitation, error) {
	inv, err := c.svc.Invitations.Create(&gclassroom.Invitation{CourseId: courseID, Role: "STUDENT", UserId: userID}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("invite student: %w", err)
	}
	return inv, nil
}

func (c *Client) RemoveTeacher(ctx context.Context, courseID, userID string) error {
	if _, err := c.svc.Courses.Teachers.Delete(courseID, userID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("remove teacher: %w", err)
	}
	return nil
}

func (c *Client) RemoveStudent(ctx context.Context, courseID, userID string) error {
	if _, err := c.svc.Courses.Students.Delete(courseID, userID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("remove student: %w", err)
	}
	return nil
}

func (c *Client) ListTopics(ctx context.Context, courseID string, params ListParams) ([]*gclassroom.Topic, string, error) {
	call := c.svc.Courses.Topics.List(courseID).Context(ctx)
	if params.PageSize > 0 {
		call = call.PageSize(params.PageSize)
	}
	if params.PageToken != "" {
		call = call.PageToken(params.PageToken)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("list topics: %w", err)
	}
	return resp.Topic, resp.NextPageToken, nil
}

func (c *Client) CreateTopic(ctx context.Context, courseID, name string) (*gclassroom.Topic, error) {
	out, err := c.svc.Courses.Topics.Create(courseID, &gclassroom.Topic{Name: name}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create topic: %w", err)
	}
	return out, nil
}

func (c *Client) UpdateTopic(ctx context.Context, courseID, topicID, name string) (*gclassroom.Topic, error) {
	out, err := c.svc.Courses.Topics.Patch(courseID, topicID, &gclassroom.Topic{Name: name}).UpdateMask("name").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("update topic: %w", err)
	}
	return out, nil
}

func (c *Client) DeleteTopic(ctx context.Context, courseID, topicID string) error {
	if _, err := c.svc.Courses.Topics.Delete(courseID, topicID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete topic: %w", err)
	}
	return nil
}

func (c *Client) MoveCourseWorkToTopic(ctx context.Context, courseID, courseWorkID, topicID string) (*gclassroom.CourseWork, error) {
	return c.PatchCourseWork(ctx, courseID, courseWorkID, &gclassroom.CourseWork{TopicId: topicID}, []string{"topicId"})
}

func (c *Client) BuildTodo(ctx context.Context) ([]TodoItem, error) {
	courses, _, err := c.ListCourses(ctx, ListParams{PageSize: 200})
	if err != nil {
		return nil, err
	}
	items := make([]TodoItem, 0)
	for _, course := range courses {
		if course.CourseState != "ACTIVE" {
			continue
		}
		courseWork, _, err := c.ListCourseWork(ctx, course.Id, ListParams{PageSize: 200})
		if err != nil {
			continue
		}
		for _, work := range courseWork {
			submissions, _, err := c.ListStudentSubmissions(ctx, course.Id, work.Id, "me", ListParams{PageSize: 1})
			if err != nil || len(submissions) == 0 {
				continue
			}
			sub := submissions[0]
			if sub.State == "TURNED_IN" || sub.State == "RETURNED" {
				continue
			}
			items = append(items, TodoItem{
				CourseID:        course.Id,
				CourseName:      course.Name,
				CourseWorkID:    work.Id,
				CourseWorkTitle: work.Title,
				DueDate:         parseDueDate(work),
				SubmissionID:    sub.Id,
				SubmissionState: sub.State,
				Late:            sub.Late,
			})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].DueDate.IsZero() {
			return false
		}
		if items[j].DueDate.IsZero() {
			return true
		}
		return items[i].DueDate.Before(items[j].DueDate)
	})
	return items, nil
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
	return time.Date(
		int(work.DueDate.Year),
		time.Month(work.DueDate.Month),
		int(work.DueDate.Day),
		hour,
		minute,
		second,
		0,
		time.Local,
	)
}
