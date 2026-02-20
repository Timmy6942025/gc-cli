# Parity Matrix

Status values:

- `Implemented`: Fully handled in API/CLI.
- `Handoff`: Not exposed in official API for full mutation parity; routed to web handoff.
- `Preview`: Supported behind preview capability gating.

| Area | Feature | Status | Notes |
| --- | --- | --- | --- |
| Classes | List/show/create/update/archive/restore/delete | Implemented | `gc classes ...` |
| Stream | Announcements list/post/edit/delete | Implemented | `gc stream ...` |
| Classwork | List/create/edit/publish/schedule/delete | Implemented | Includes topic and due-date support |
| Classwork | Attachment upload/Drive/link materials | Implemented | Drive API-backed |
| Submissions | List/show/turn in/reclaim/unsubmit | Implemented | Student actions |
| Submissions | Draft/assigned grading + return | Implemented | Teacher actions |
| People | List/invite/remove teacher/student | Implemented | Classroom roster + invitation APIs |
| Topics | List/create/edit/delete/move classwork | Implemented | `gc topics ...` |
| To-do | Student to-do aggregation | Implemented | Derived from coursework + `me` submissions |
| Calendar | Open Google Calendar | Implemented | Web open action |
| Settings | Invite code reset/disable | Handoff | `gc handoff open invite-code --course ...` |
| Settings | Stream moderation controls | Handoff | Classroom web-only |
| Settings | Meet link controls | Handoff | Classroom web-only |
| Analytics | Classroom analytics | Handoff | Classroom web-only |
| Preview | Rubrics | Preview | `preview_enabled` gate |
| Preview | Grading periods | Preview | `preview_enabled` gate |
| Preview | Student groups | Preview | `preview_enabled` gate |
