package domain

type UserResponse struct {
    Uuid      string   `json:"uuid"`
    Email     string   `json:"email"`
    UserName  string   `json:"user_name"`
    Avatar    string   `json:"avatar"`
    AboutMe   string   `json:"about_me"`
    Friends   []string `json:"friends"`
    Status    string   `json:"status"`
    CreatedAt string   `json:"created_at"`
    UpdatedAt string   `json:"updated_at"`
}

