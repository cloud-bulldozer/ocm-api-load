package elastic

import (
	"os"
	"reflect"
	"testing"
)

func Test_extractENVContent(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want map[string]string
	}{
		{"manual_run", map[string]string{}, map[string]string{"buildUrl": "Manual run", "CiSystem": "Local"}},
		{"has_build_url", map[string]string{"BUILD_URL": "https://test.com"}, map[string]string{"buildUrl": "https://test.com", "CiSystem": "Local"}},
		{"jenkins_run", map[string]string{"BUILD_URL": "https://test.com", "JENKINS_URL": "http://jenkins.dev"}, map[string]string{"buildUrl": "https://test.com", "CiSystem": "Jenkins"}},
		{"airflow_run", map[string]string{"BUILD_URL": "https://test.com", "AIRFLOW_CTX_DAG_ID": "ocm-dag"}, map[string]string{"buildUrl": "https://test.com", "CiSystem": "Airflow"}},
		{"jenkins_run", map[string]string{"BUILD_URL": "https://test.com", "JENKINS_URL": "http://jenkins.dev", "AIRFLOW_CTX_DAG_ID": "ocm-dag"}, map[string]string{"buildUrl": "https://test.com", "CiSystem": "Jenkins"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, v := range tt.envs {
				os.Setenv(key, v)
			}
			if got := extractENVContent(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractENVContent() = %v, want %v", got, tt.want)
			}
			for key := range tt.envs {
				os.Unsetenv(key)
			}
		})
	}
}

// buildURL := os.Getenv("BUILD_URL")
// if buildURL != "" {
// 	result["buildUrl"] = buildURL
// } else {
// 	result["buildUrl"] = "Manual run"
// }
// dagID := os.Getenv("AIRFLOW_CTX_DAG_ID")
// jenkinsURL := os.Getenv("JENKINS_URL")
