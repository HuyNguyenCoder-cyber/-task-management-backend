package dto

type AddProjectMemberRequest struct {
	Email string `json:"email" binding:"required"`
}

type ProjectMemberResponse struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}
