package rn

import (
	"errors"
	"testing"
)

func TestParseRN(t *testing.T) {
	cases := []struct {
		input string
		arn   RN
		err   error
	}{
		{
			input: "invalid",
			err:   errors.New(invalidPrefix),
		},
		{
			input: "arn:nope",
			err:   errors.New(invalidSections),
		},
		{
			input: "rn:nope",
			err:   errors.New(invalidSections),
		},
		{
			input: "rn::nats:::tests.create",
			arn: RN{
				Service:  "nats",
				Resource: "tests.create",
			},
		},
		{
			input: "arn:aws:sns:ap-northeast-1:123456789012:test",
			arn: RN{
				Service:  "sns",
				Resource: "test",
			},
		},
		{
			input: "arn:aws:sns:ap-northeast-1:123456789012:test.fifo",
			arn: RN{
				Service:  "sns",
				Resource: "test.fifo",
			},
		},
		{
			input: "arn:aws:ecr:us-west-2:123456789012:repository/foo/bar",
			arn: RN{
				Service:  "ecr",
				Resource: "repository/foo/bar",
			},
		},
		{
			input: "arn:aws:elasticbeanstalk:us-east-1:123456789012:environment/My App/MyEnvironment",
			arn: RN{
				Service:  "elasticbeanstalk",
				Resource: "environment/My App/MyEnvironment",
			},
		},
		{
			input: "arn:aws:iam::123456789012:user/David",
			arn: RN{
				Service:  "iam",
				Resource: "user/David",
			},
		},
		{
			input: "arn:aws:rds:eu-west-1:123456789012:db:mysql-db",
			arn: RN{
				Service:  "rds",
				Resource: "db:mysql-db",
			},
		},
		{
			input: "arn:aws:s3:::my_corporate_bucket/exampleobject.png",
			arn: RN{
				Service:  "s3",
				Resource: "my_corporate_bucket/exampleobject.png",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			spec, err := Parse(tc.input)
			if tc.arn != spec {
				t.Errorf("Expected %q to parse as %v, but got %v", tc.input, tc.arn, spec)
			}
			if err == nil && tc.err != nil {
				t.Errorf("Expected err to be %v, but got nil", tc.err)
			} else if err != nil && tc.err == nil {
				t.Errorf("Expected err to be nil, but got %v", err)
			} else if err != nil && tc.err != nil && err.Error() != tc.err.Error() {
				t.Errorf("Expected err to be %v, but got %v", tc.err, err)
			}
		})
	}
}

func TestIsRN(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		In     string
		Expect bool
		// Params
	}{
		"valid ARN slash resource": {
			In:     "arn:aws:service:us-west-2:123456789012:restype/resvalue",
			Expect: true,
		},
		"valid ARN colon resource": {
			In:     "arn:aws:service:us-west-2:123456789012:restype:resvalue",
			Expect: true,
		},
		"valid ARN resource": {
			In:     "arn:aws:service:us-west-2:123456789012:*",
			Expect: true,
		},
		"valid RN resource": {
			In:     "rn::service:::*",
			Expect: true,
		},
		"empty sections(arn)": {
			In:     "arn:::::",
			Expect: true,
		},
		"empty sections(rn)": {
			In:     "rn:::::",
			Expect: true,
		},
		"invalid ARN": {
			In: "some random string",
		},
		"invalid ARN missing resource": {
			In: "arn:aws:service:us-west-2:123456789012",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual := IsRN(c.In)
			if e, a := c.Expect, actual; e != a {
				t.Errorf("expect %s valid %v, got %v", c.In, e, a)
			}
		})
	}
}
