package clerk_user_center

// UserCenterConfig 描述在 Answer 后台插件页面需要填写的配置项
type UserCenterConfig struct {
    PublishableKey string `json:"publishable_key" title:"Clerk Publishable Key" description:"用于前端初始化 Clerk.js 的 publishableKey" required:"true"`
    SecretKey      string `json:"secret_key"      title:"Clerk Secret Key"      description:"用于后端验证 session 的 Secret Key"      required:"true"`
    FrontendAPI    string `json:"frontend_api"    title:"Clerk Frontend API"    description:"Clerk 前端 API 地址，例如：https://api.clerk.dev" required:"true"`
}
