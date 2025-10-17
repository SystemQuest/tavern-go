package fixtures

import (
	"github.com/systemquest/tavern-go/pkg/schema"
)

// CreateSimpleTest creates a simple single-stage test specification
func CreateSimpleTest(url, method string, expectedStatus int) *schema.TestSpec {
	return &schema.TestSpec{
		TestName: "Simple test",
		Stages: []schema.Stage{
			{
				Name: "Single stage",
				Request: schema.RequestSpec{
					URL:    url,
					Method: method,
				},
				Response: schema.ResponseSpec{
					StatusCode: expectedStatus,
				},
			},
		},
	}
}

// CreateMultiStageTest creates a multi-stage test specification
func CreateMultiStageTest(testName string, stages []schema.Stage) *schema.TestSpec {
	return &schema.TestSpec{
		TestName: testName,
		Stages:   stages,
	}
}

// CreateTestWithVariables creates a test with variable substitution
func CreateTestWithVariables(url, method string, vars map[string]interface{}) *schema.TestSpec {
	return &schema.TestSpec{
		TestName: "Test with variables",
		Stages: []schema.Stage{
			{
				Name: "Stage with variables",
				Request: schema.RequestSpec{
					URL:    url,
					Method: method,
					Headers: map[string]string{
						"X-Test-Header": "{test_var}",
					},
				},
				Response: schema.ResponseSpec{
					StatusCode: 200,
				},
			},
		},
	}
}

// CreateAuthTest creates a test with authentication
func CreateAuthTest(url, token string) *schema.TestSpec {
	return &schema.TestSpec{
		TestName: "Auth test",
		Stages: []schema.Stage{
			{
				Name: "Get token",
				Request: schema.RequestSpec{
					URL:    url + "/auth",
					Method: "POST",
					JSON: map[string]interface{}{
						"username": "test",
						"password": "secret",
					},
				},
				Response: schema.ResponseSpec{
					StatusCode: 200,
					Save: &schema.SaveSpec{
						Body: map[string]string{
							"token": "access_token",
						},
					},
				},
			},
			{
				Name: "Use token",
				Request: schema.RequestSpec{
					URL:    url + "/protected",
					Method: "GET",
					Headers: map[string]string{
						"Authorization": "Bearer {token}",
					},
				},
				Response: schema.ResponseSpec{
					StatusCode: 200,
					Body: map[string]interface{}{
						"authenticated": true,
					},
				},
			},
		},
	}
}

// CreateTestWithSave creates a test that saves response values
func CreateTestWithSave(url string, saveKey string, savePath string) *schema.TestSpec {
	return &schema.TestSpec{
		TestName: "Test with save",
		Stages: []schema.Stage{
			{
				Name: "Get data",
				Request: schema.RequestSpec{
					URL:    url,
					Method: "GET",
				},
				Response: schema.ResponseSpec{
					StatusCode: 200,
					Save: &schema.SaveSpec{
						Body: map[string]string{
							saveKey: savePath,
						},
					},
				},
			},
		},
	}
}
