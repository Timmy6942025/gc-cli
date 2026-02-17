package auth

import (
	"sort"
)

const (
	ScopeOpenID                      = "openid"
	ScopeUserEmail                   = "https://www.googleapis.com/auth/userinfo.email"
	ScopeUserProfile                 = "https://www.googleapis.com/auth/userinfo.profile"
	ScopeCoursesReadonly             = "https://www.googleapis.com/auth/classroom.courses.readonly"
	ScopeCourses                     = "https://www.googleapis.com/auth/classroom.courses"
	ScopeRostersReadonly             = "https://www.googleapis.com/auth/classroom.rosters.readonly"
	ScopeRosters                     = "https://www.googleapis.com/auth/classroom.rosters"
	ScopeAnnouncementsReadonly       = "https://www.googleapis.com/auth/classroom.announcements.readonly"
	ScopeAnnouncements               = "https://www.googleapis.com/auth/classroom.announcements"
	ScopeCourseWorkMeReadonly        = "https://www.googleapis.com/auth/classroom.coursework.me.readonly"
	ScopeCourseWorkMe                = "https://www.googleapis.com/auth/classroom.coursework.me"
	ScopeCourseWorkStudentsReadonly  = "https://www.googleapis.com/auth/classroom.coursework.students.readonly"
	ScopeCourseWorkStudents          = "https://www.googleapis.com/auth/classroom.coursework.students"
	ScopeTopicsReadonly              = "https://www.googleapis.com/auth/classroom.topics.readonly"
	ScopeTopics                      = "https://www.googleapis.com/auth/classroom.topics"
	ScopeCourseWorkMaterialsReadonly = "https://www.googleapis.com/auth/classroom.courseworkmaterials.readonly"
	ScopeCourseWorkMaterials         = "https://www.googleapis.com/auth/classroom.courseworkmaterials"
	ScopeDriveFile                   = "https://www.googleapis.com/auth/drive.file"
)

var DefaultReadScopes = []string{
	ScopeOpenID,
	ScopeUserEmail,
	ScopeUserProfile,
	ScopeCoursesReadonly,
	ScopeRostersReadonly,
	ScopeAnnouncementsReadonly,
	ScopeCourseWorkMeReadonly,
	ScopeCourseWorkStudentsReadonly,
	ScopeTopicsReadonly,
	ScopeCourseWorkMaterialsReadonly,
}

var DefaultWriteScopes = []string{
	ScopeOpenID,
	ScopeUserEmail,
	ScopeUserProfile,
	ScopeCourses,
	ScopeRosters,
	ScopeAnnouncements,
	ScopeCourseWorkMe,
	ScopeCourseWorkStudents,
	ScopeTopics,
	ScopeCourseWorkMaterials,
	ScopeDriveFile,
}

func NormalizeScopes(scopes []string) []string {
	uniq := map[string]struct{}{}
	for _, s := range scopes {
		if s == "" {
			continue
		}
		uniq[s] = struct{}{}
	}
	out := make([]string, 0, len(uniq))
	for s := range uniq {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
