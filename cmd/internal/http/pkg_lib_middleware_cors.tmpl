package middleware

import "net/http"

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()

		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Credentials", "true")
		header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, withCredentials")
		// header.Set("Access-Control-Expose-Headers", "服务器暴露一些自定义的头信息，允许客户端访问")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
