package handler

import "testing"

func TestValidateScSmartPickReq(t *testing.T) {
	tests := []struct {
		name    string
		req     scSmartPickReq
		wantErr bool
	}{
		{name: "both zero", req: scSmartPickReq{}, wantErr: true},
		{name: "negative JAV", req: scSmartPickReq{JAV: scSmartPickJAVReq{PickN: -1}, FC2: scSmartPickFC2Req{PickN: 1, Reqs: []scPickReq{{Weight: 1}}}}, wantErr: true},
		{name: "negative FC2", req: scSmartPickReq{JAV: scSmartPickJAVReq{PickN: 1, Reqs: []scPickReq{{Weight: 1}}}, FC2: scSmartPickFC2Req{PickN: -1}}, wantErr: true},
		{name: "JAV reqs missing", req: scSmartPickReq{JAV: scSmartPickJAVReq{PickN: 1}}, wantErr: true},
		{name: "FC2 reqs missing", req: scSmartPickReq{FC2: scSmartPickFC2Req{PickN: 1}}, wantErr: true},
		{name: "JAV only", req: scSmartPickReq{JAV: scSmartPickJAVReq{PickN: 2, Reqs: []scPickReq{{Weight: 1}}}}},
		{name: "FC2 only", req: scSmartPickReq{FC2: scSmartPickFC2Req{PickN: 3, Reqs: []scPickReq{{Weight: 1}}}}},
		{name: "mixed", req: scSmartPickReq{
			JAV: scSmartPickJAVReq{PickN: 2, Reqs: []scPickReq{{Weight: 1}}},
			FC2: scSmartPickFC2Req{PickN: 3, Reqs: []scPickReq{{Weight: 1}}},
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateScSmartPickReq(test.req)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateScSmartPickReq() error = %v, wantErr = %v", err, test.wantErr)
			}
		})
	}
}

func TestFormatSmartPickSizeGB(t *testing.T) {
	if got := formatSmartPickSizeGB(3 * 1024 * 1024 * 1024); got != "3.00" {
		t.Fatalf("formatSmartPickSizeGB() = %q", got)
	}
}
