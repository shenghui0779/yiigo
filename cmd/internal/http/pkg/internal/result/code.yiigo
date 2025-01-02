package result

func OK(options ...Option) Result {
	return New(CodeOK, "OK", options...)
}

func ErrParams(options ...Option) Result {
	return New(10000, "参数错误", options...)
}

func ErrAuth(options ...Option) Result {
	return New(20000, "未授权，请先登录", options...)
}

func ErrPerm(options ...Option) Result {
	return New(30000, "权限不足", options...)
}

func ErrNotFound(options ...Option) Result {
	return New(40000, "数据不存在", options...)
}

func ErrSystem(options ...Option) Result {
	return New(50000, "内部服务器错误，请稍后重试", options...)
}

func ErrData(options ...Option) Result {
	return New(60000, "数据异常，请稍后重试", options...)
}

func ErrService(options ...Option) Result {
	return New(70000, "服务异常，请稍后重试", options...)
}
