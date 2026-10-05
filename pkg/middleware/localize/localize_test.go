package localize

import (
	"context"
	"testing"
)

func TestTranslate(t *testing.T) {
	cases := []struct {
		name string
		lang string
		key  string
		want string
	}{
		{"zh invalid credentials", "zh-cn", "INVALID_CREDENTIALS", "用户名或密码错误"},
		{"en invalid credentials", "en-us", "INVALID_CREDENTIALS", "Invalid username or password"},
		{"zh permission denied", "zh-cn", "PERMISSION_DENIED", "无接口访问权限"},
		{"en permission denied", "en-us", "PERMISSION_DENIED", "No permission to access this API"},
		{"zh validate dictType.name.required", "zh-cn", "dictType.name.required", "名称不能为空"},
		{"miss key fallback", "zh-cn", "NOT_EXIST_KEY", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), localizerKey{}, c.lang)
			got, ok := Translate(ctx, c.key)
			if c.want == "" {
				if ok {
					t.Fatalf("expected miss, got %q", got)
				}
				return
			}
			if !ok || got != c.want {
				t.Fatalf("got %q ok=%v, want %q", got, ok, c.want)
			}
		})
	}
}

func TestNormalizeLang(t *testing.T) {
	cases := map[string]string{
		"zh-CN":          "zh-cn",
		"en-US":          "en-us",
		"zh-CN,zh;q=0.9": "zh-cn",
		"en-US,en;q=0.8": "en-us",
		"zh-Hans":        "zh-cn",
		"":               "",
	}
	for in, want := range cases {
		if got := normalizeLang(in); got != want {
			t.Fatalf("normalizeLang(%q) = %q, want %q", in, got, want)
		}
	}
}
