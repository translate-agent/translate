package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
)

func Test_FromArb(t *testing.T) {
	t.Parallel()

	type args struct {
		in0 []byte
	}

	tests := []struct {
		name    string
		args    args
		want    model.Messages
		wantErr bool
	}{
		{
			name: "a",
			args: args{
				in0: []byte(`
				{
					"title": "Hello World!",
					"@title" : {
						"description" : "Message to greet the World"
					},
					"greeting": "Welcome {user}!",
					"@greeting": {
						"placeholders": {
							"user":{
								"type":"string",
								"example":"Bob"
							}
						}
					},
					"aaa":"sss" 
				}`),
			},
			want: model.Messages{
				Messages: []model.Message{
					{
						ID:          "title",
						Message:     "Hello World!",
						Description: "Message to greet the World",
					},
					{
						ID:      "greeting",
						Message: "Welcome {user}!",
					},
					{
						ID:      "aaa",
						Message: "sss",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res, err := FromArb(tt.args.in0)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, res)
		})
	}
}

func Test_ToArb(t *testing.T) {
	t.Parallel()

	messages := model.Messages{
		Messages: []model.Message{
			{
				ID:          "title",
				Message:     "Hello World!",
				Description: "Message to greet the World",
			},
			{
				ID:      "greeting",
				Message: "Welcome {user}",
			},
		},
	}

	want := []byte(`
	{
		"title":"Hello World!",
		"@title":{
			"description":"Message to greet the World"
		},
		"greeting":"Welcome {user}"
	}
	`)

	res, err := ToArb(messages)
	if !assert.NoError(t, err) {
		return
	}

	assert.JSONEq(t, string(want), string(res))

	// type args struct {
	// 	messages model.Messages
	// }

	// tests := []struct {
	// 	name    string
	// 	args    args
	// 	want    []byte
	// 	wantErr bool
	// }{
	// 	{
	// 		name: "Messages to ARB",
	// 		args: args{
	// 			messages: model.Messages{
	// 				Messages: []model.Message{
	// 					{
	// 						ID:          "title",
	// 						Message:     "Hello World!",
	// 						Description: "Message to greet the World",
	// 					},
	// 					{
	// 						ID:      "greeting",
	// 						Message: "Welcome {user}",
	// 					},
	// 				},
	// 			},
	// 		},
	// 		want: []byte(`
	// 		{
	// 			"title":"Hello World!",
	// 			"@title":{
	// 				"description":"Message to greet the World"
	// 			},
	// 			"greeting":"Welcome {user}"
	// 		}
	// 		`),
	// 	},
	// 	{
	// 		name: "With Description",
	// 		args: args{
	// 			messages: model.Messages{
	// 				Messages: []model.Message{
	// 					{
	// 						ID:          "title",
	// 						Message:     "Hello World!",
	// 						Description: "Message to greet the World",
	// 					},
	// 					{
	// 						ID:          "greeting",
	// 						Message:     "Welcome {user}",
	// 						Description: "Welcome user to web page",
	// 					},
	// 				},
	// 			},
	// 		},
	// 		want: []byte(`
	// 		{
	// 			"title":"Hello World!",
	// 			"@title":{
	// 				"description":"Message to greet the World"
	// 			},
	// 			"greeting":"Welcome {user}"
	// 		}
	// 		`),
	// 	},
	// }
	// for _, tt := range tests {
	// 	tt := tt
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		t.Parallel()

	// 		res, err := ToArb(tt.args.messages)
	// 		if !assert.NoError(t, err) {
	// 			return
	// 		}

	// 		assert.JSONEq(t, string(tt.want), string(res))
	// 	})
	// }
}
