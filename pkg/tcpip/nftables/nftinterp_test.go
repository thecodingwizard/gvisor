// Copyright 2024 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nftables

import (
	"fmt"
	"testing"

	"gvisor.dev/gvisor/pkg/abi/linux"
)

type interpretOperationTestAction struct {
	tname    string
	opStr    string
	expected operation // will be nil if an error is expected
}

// checkOp is a generic operation validation function used for testing that
// the interpretation of an operation matches the expected operation.
func checkOp(t *testing.T, test interpretOperationTestAction, checkFunc func(string, operation, operation) error) {
	rule, err := InterpretRule(test.opStr)
	if test.expected == nil {
		if err == nil {
			t.Fatalf("unexpected interpretation success for %s", test.tname)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected interpretation error for %s: %v", test.tname, err)
	}
	if len(rule.ops) != 1 {
		t.Fatalf("expected single operation for %s, got %d", test.tname, len(rule.ops))
	}
	actual := rule.ops[0]
	if actual == nil {
		t.Fatalf("expected non-nil operation for %s, got nil", test.tname)
	}
	if err := checkFunc(test.tname, test.expected, actual); err != nil {
		t.Fatalf(err.Error())
	}
}

// TestInterpretImmediateOps tests interpretation of immediate operations.
func TestInterpretImmediateOps(t *testing.T) {
	for _, test := range []interpretOperationTestAction{
		{
			tname:    "verdict register with accept verdict",
			opStr:    "[ immediate reg 0 accept ]",
			expected: mustCreateImmediate(t, linux.NFT_REG_VERDICT, newVerdictData(Verdict{Code: VC(linux.NF_ACCEPT)})),
		},
		{
			tname:    "verdict register with drop verdict",
			opStr:    "[ immediate reg 0 drop ]",
			expected: mustCreateImmediate(t, linux.NFT_REG_VERDICT, newVerdictData(Verdict{Code: VC(linux.NF_DROP)})),
		},
		{
			tname:    "verdict register with continue verdict",
			opStr:    "[ immediate reg 0 continue ]",
			expected: mustCreateImmediate(t, linux.NFT_REG_VERDICT, newVerdictData(Verdict{Code: VC(linux.NFT_CONTINUE)})),
		},
		{
			tname:    "verdict register with return verdict",
			opStr:    "[ immediate reg 0 return ]",
			expected: mustCreateImmediate(t, linux.NFT_REG_VERDICT, newVerdictData(Verdict{Code: VC(linux.NFT_RETURN)})),
		},
		{
			tname:    "verdict register with jump verdict",
			opStr:    "[ immediate reg 0 jump -> next_chain ]",
			expected: mustCreateImmediate(t, linux.NFT_REG_VERDICT, newVerdictData(Verdict{Code: VC(linux.NFT_JUMP), ChainName: "next_chain"})),
		},
		{
			tname:    "verdict register with goto verdict",
			opStr:    "[ immediate reg 0 goto -> next_chain ]",
			expected: mustCreateImmediate(t, linux.NFT_REG_VERDICT, newVerdictData(Verdict{Code: VC(linux.NFT_GOTO), ChainName: "next_chain"})),
		},
		{
			tname:    "verdict register with 4-byte data",
			opStr:    "[ immediate reg 0 0x0201a8c0 ]",
			expected: nil,
		},
		{
			tname:    "verdict register with 16-byte data",
			opStr:    "[ immediate reg 0 0xb80d0120 0x00000000 0x00000000 0x02000000 ]",
			expected: nil,
		},
		{
			tname:    "16-byte register with verdict data",
			opStr:    "[ immediate reg 1 accept ]",
			expected: nil,
		},
		{
			tname:    "16-byte register with verdict data with target",
			opStr:    "[ immediate reg 2 jump -> next_chain ]",
			expected: nil,
		},
		{
			tname:    "16-byte register with 4-byte data",
			opStr:    "[ immediate reg 1 0x0201a8c0 ]",
			expected: mustCreateImmediate(t, linux.NFT_REG_1, newBytesData([]byte{0xc0, 0xa8, 0x01, 0x02})),
		},
		{
			tname:    "16-byte register with 8-byte data",
			opStr:    "[ immediate reg 2 0xb80d0120 0x00000050 ]",
			expected: mustCreateImmediate(t, linux.NFT_REG_2, newBytesData([]byte{0x20, 0x01, 0x0d, 0xb8, 0x50, 0x00, 0x00, 0x00})),
		},
		{
			tname:    "16-byte register with 12-byte data",
			opStr:    "[ immediate reg 3 0xb80d0120 0x00000050 0xb80d0120 ]",
			expected: mustCreateImmediate(t, linux.NFT_REG_3, newBytesData([]byte{0x20, 0x01, 0x0d, 0xb8, 0x50, 0x00, 0x00, 0x00, 0x20, 0x01, 0x0d, 0xb8})),
		},
		{
			tname:    "16-byte register with 16-byte data",
			opStr:    "[ immediate reg 4 0xb80d0120 0x00000000 0x00000000 0x02000000 ]",
			expected: mustCreateImmediate(t, linux.NFT_REG_4, newBytesData([]byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02})),
		},
		{
			tname:    "4-byte register with verdict data",
			opStr:    "[ immediate reg 8 return ]",
			expected: nil,
		},
		{
			tname:    "4-byte register with verdict data with target",
			opStr:    "[ immediate reg 9 goto -> next_chain ]",
			expected: nil,
		},
		{
			tname:    "4-byte register with 16-byte data",
			opStr:    "[ immediate reg 9 0xb80d0120 0x00000000 0x00000000 0x02000000 ]",
			expected: nil,
		},
	} {
		t.Run(test.tname, func(t *testing.T) { checkOp(t, test, checkImmediateOp) })
	}
}

// checkImmediateOp checks that the given operation is an immediate operation
// and that it matches the expected immediate operation.
func checkImmediateOp(tname string, expected operation, actual operation) error {
	expectedImm := expected.(*immediate)
	imm, ok := actual.(*immediate)
	if !ok {
		return fmt.Errorf("expected operation type to be Immediate for %s, got %T", tname, actual)
	}
	if imm.dreg != expectedImm.dreg {
		return fmt.Errorf("expected register to be %d for %s, got %d", expectedImm.dreg, tname, imm.dreg)
	}
	if !imm.data.equal(expectedImm.data) {
		return fmt.Errorf("expected data to be %v for %s, got %v", expectedImm.data, tname, imm.data)
	}
	return nil
}

// TestInterpretComparisonOps tests interpretation of comparison operations.
func TestInterpretComparisonOps(t *testing.T) {
	for _, test := range []interpretOperationTestAction{
		{
			tname:    "verdict register with 4-byte data comparison",
			opStr:    "[ cmp eq reg 0 0x00000002 ]",
			expected: nil,
		},
		{
			tname:    "verdict register with 8-byte data comparison",
			opStr:    "[ cmp lt reg 0 0xb80d0120 0x02000000 ]",
			expected: nil,
		},
		{
			tname:    "verdict register with 12-byte data comparison",
			opStr:    "[ cmp gte reg 0 0xb80d0120 0x18305290 0x02000000 ]",
			expected: nil,
		},
		{
			tname:    "verdict register with 16-byte data comparison",
			opStr:    "[ cmp neq reg 0 0xb80d0120 0x18305290 0x18305290 0x02000000 ]",
			expected: nil,
		},
		{
			tname:    "4-byte register == 4-byte data",
			opStr:    "[ cmp eq reg 8 0x0302010a ]",
			expected: mustCreateComparison(t, linux.NFT_REG32_00, linux.NFT_CMP_EQ, newBytesData([]byte{0x0a, 0x01, 0x02, 0x03})),
		},
		{
			tname:    "4-byte register != 4-byte data",
			opStr:    "[ cmp neq reg 9 0x00000064 ]",
			expected: mustCreateComparison(t, linux.NFT_REG32_01, linux.NFT_CMP_NEQ, newBytesData([]byte{0x64, 0x00, 0x00, 0x00})),
		},
		{
			tname:    "4-byte register < 4-byte data",
			opStr:    "[ cmp lt reg 10 0x00000000 ]",
			expected: mustCreateComparison(t, linux.NFT_REG32_02, linux.NFT_CMP_LT, newBytesData([]byte{0x00, 0x00, 0x00, 0x00})),
		},
		{
			tname:    "4-byte register <= 4-byte data",
			opStr:    "[ cmp lte reg 11 0x00000164 ]",
			expected: mustCreateComparison(t, linux.NFT_REG32_03, linux.NFT_CMP_LTE, newBytesData([]byte{0x64, 0x01, 0x00, 0x00})),
		},
		{
			tname:    "4-byte register > 4-byte data",
			opStr:    "[ cmp gt reg 12 0xe8030000 ]",
			expected: mustCreateComparison(t, linux.NFT_REG32_04, linux.NFT_CMP_GT, newBytesData([]byte{0x00, 0x00, 0x03, 0xe8})),
		},
		{
			tname:    "4-byte register >= 4-byte data",
			opStr:    "[ cmp gte reg 13 0xc02b0000 ]",
			expected: mustCreateComparison(t, linux.NFT_REG32_05, linux.NFT_CMP_GTE, newBytesData([]byte{0x00, 0x00, 0x2b, 0xc0})),
		},
		{
			tname:    "4-byte register with 8-byte data comparison",
			opStr:    "[ cmp eq reg 14 0xb80d0120 0x02000000 ]",
			expected: nil,
		},
		{
			tname:    "4-byte register with 12-byte data comparison",
			opStr:    "[ cmp lte reg 15 0xb80d0120 0x18305290 0x02000000 ]",
			expected: nil,
		},
		{
			tname:    "4-byte register with 16-byte data comparison",
			opStr:    "[ cmp gt reg 16 0x0302010a 0x00000000 0x00000000 0x02000001 ]",
			expected: nil,
		},
		{
			tname:    "16-byte register == 4-byte data",
			opStr:    "[ cmp eq reg 1 0x0302010a ]",
			expected: mustCreateComparison(t, linux.NFT_REG_1, linux.NFT_CMP_EQ, newBytesData([]byte{0x0a, 0x01, 0x02, 0x03})),
		},
		{
			tname:    "16-byte register != 4-byte data",
			opStr:    "[ cmp neq reg 2 0x00000064 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_2, linux.NFT_CMP_NEQ, newBytesData([]byte{0x64, 0x00, 0x00, 0x00})),
		},
		{
			tname:    "16-byte register < 4-byte data",
			opStr:    "[ cmp lt reg 3 0x00000000 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_3, linux.NFT_CMP_LT, newBytesData([]byte{0x00, 0x00, 0x00, 0x00})),
		},
		{
			tname:    "16-byte register <= 4-byte data",
			opStr:    "[ cmp lte reg 4 0x00000164 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_4, linux.NFT_CMP_LTE, newBytesData([]byte{0x64, 0x01, 0x00, 0x00})),
		},
		{
			tname:    "16-byte register > 4-byte data",
			opStr:    "[ cmp gt reg 1 0xe8030000 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_1, linux.NFT_CMP_GT, newBytesData([]byte{0x00, 0x00, 0x03, 0xe8})),
		},
		{
			tname:    "16-byte register >= 4-byte data",
			opStr:    "[ cmp gte reg 2 0xc02b0000 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_2, linux.NFT_CMP_GTE, newBytesData([]byte{0x00, 0x00, 0x2b, 0xc0})),
		},
		{
			tname:    "16-byte register == 8-byte data",
			opStr:    "[ cmp eq reg 1 0x0302010a 0x12345678 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_1, linux.NFT_CMP_EQ, newBytesData([]byte{0x0a, 0x01, 0x02, 0x03, 0x78, 0x56, 0x34, 0x12})),
		},
		{
			tname:    "16-byte register != 8-byte data",
			opStr:    "[ cmp neq reg 2 0x00000064 0x00000020 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_2, linux.NFT_CMP_NEQ, newBytesData([]byte{0x64, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00})),
		},
		{
			tname:    "16-byte register < 8-byte data",
			opStr:    "[ cmp lt reg 3 0x00000000 0x00000000 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_3, linux.NFT_CMP_LT, newBytesData([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})),
		},
		{
			tname:    "16-byte register <= 8-byte data",
			opStr:    "[ cmp lte reg 4 0x00000164 0x00000164 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_4, linux.NFT_CMP_LTE, newBytesData([]byte{0x64, 0x01, 0x00, 0x00, 0x64, 0x01, 0x00, 0x00})),
		},
		{
			tname:    "16-byte register > 8-byte data",
			opStr:    "[ cmp gt reg 2 0xe8030000 0x00000f13 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_2, linux.NFT_CMP_GT, newBytesData([]byte{0x00, 0x00, 0x03, 0xe8, 0x13, 0x0f, 0x00, 0x00})),
		},
		{
			tname:    "16-byte register >= 8-byte data",
			opStr:    "[ cmp gte reg 3 0x0a000120 0xc0090000 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_3, linux.NFT_CMP_GTE, newBytesData([]byte{0x20, 0x01, 0x00, 0x0a, 0x00, 0x00, 0x09, 0xc0})),
		},
		{
			tname:    "16-byte register == 12-byte data",
			opStr:    "[ cmp eq reg 1 0x0302010a 0x00000000 0x12345678 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_1, linux.NFT_CMP_EQ, newBytesData([]byte{0x0a, 0x01, 0x02, 0x03, 0x00, 0x00, 0x00, 0x00, 0x78, 0x56, 0x34, 0x12})),
		},
		{
			tname:    "16-byte register != 12-byte data",
			opStr:    "[ cmp neq reg 2 0x00000064 0x00000000 0x00000020 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_2, linux.NFT_CMP_NEQ, newBytesData([]byte{0x64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00})),
		},
		{
			tname:    "16-byte register < 12-byte data",
			opStr:    "[ cmp lt reg 3 0x00000000 0x00000000 0x00000000 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_3, linux.NFT_CMP_LT, newBytesData([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})),
		},
		{
			tname:    "16-byte register <= 12-byte data",
			opStr:    "[ cmp lte reg 4 0x00000164 0x00000164 0x00000164 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_4, linux.NFT_CMP_LTE, newBytesData([]byte{0x64, 0x01, 0x00, 0x00, 0x64, 0x01, 0x00, 0x00, 0x64, 0x01, 0x00, 0x00})),
		},
		{
			tname:    "16-byte register > 12-byte data",
			opStr:    "[ cmp gt reg 2 0xe8030000 0x00000f13 0xc0090000 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_2, linux.NFT_CMP_GT, newBytesData([]byte{0x00, 0x00, 0x03, 0xe8, 0x13, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x09, 0xc0})),
		},
		{
			tname:    "16-byte register >= 12-byte data",
			opStr:    "[ cmp gte reg 3 0x0a000120 0x00000f13 0xc0090000 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_3, linux.NFT_CMP_GTE, newBytesData([]byte{0x20, 0x01, 0x00, 0x0a, 0x13, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x09, 0xc0})),
		},
		{
			tname:    "16-byte register == 16-byte data",
			opStr:    "[ cmp eq reg 1 0x0302010a 0x00000000 0x00000000 0x02000002 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_1, linux.NFT_CMP_EQ, newBytesData([]byte{0x0a, 0x01, 0x02, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x02})),
		},
		{
			tname:    "16-byte register != 16-byte data",
			opStr:    "[ cmp neq reg 2 0x00000064 0x00000000 0x00000000 0x02000000 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_2, linux.NFT_CMP_NEQ, newBytesData([]byte{0x64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02})),
		},
		{
			tname:    "16-byte register < 16-byte data",
			opStr:    "[ cmp lt reg 3 0x00000000 0x00000000 0x00000000 0x00000000 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_3, linux.NFT_CMP_LT, newBytesData([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})),
		},
		{
			tname:    "16-byte register <= 16-byte data",
			opStr:    "[ cmp lte reg 4 0x00000164 0x00000164 0x00000164 0x00000164 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_4, linux.NFT_CMP_LTE, newBytesData([]byte{0x64, 0x01, 0x00, 0x00, 0x64, 0x01, 0x00, 0x00, 0x64, 0x01, 0x00, 0x00, 0x64, 0x01, 0x00, 0x00})),
		},
		{
			tname:    "16-byte register > 16-byte data",
			opStr:    "[ cmp gt reg 2 0xe8030000 0x00000f13 0xc0090000 0x0b136a87 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_2, linux.NFT_CMP_GT, newBytesData([]byte{0x00, 0x00, 0x03, 0xe8, 0x13, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x09, 0xc0, 0x87, 0x6a, 0x13, 0x0b})),
		},
		{
			tname:    "16-byte register >= 16-byte data",
			opStr:    "[ cmp gte reg 3 0x0a000120 0x00000f13 0xc0090000 0x0b136a87 ]",
			expected: mustCreateComparison(t, linux.NFT_REG_3, linux.NFT_CMP_GTE, newBytesData([]byte{0x20, 0x01, 0x00, 0x0a, 0x13, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x09, 0xc0, 0x87, 0x6a, 0x13, 0x0b})),
		},
	} {
		t.Run(test.tname, func(t *testing.T) { checkOp(t, test, checkComparisonOp) })
	}
}

// checkComparisonOp checks that the given operation is an comparison operation
// and that it matches the expected comparison operation.
func checkComparisonOp(tname string, expected operation, actual operation) error {
	expectedCmp := expected.(*comparison)
	cmp, ok := actual.(*comparison)
	if !ok {
		return fmt.Errorf("expected operation type to be Comparison for %s, got %T", tname, actual)
	}
	if cmp.sreg != expectedCmp.sreg {
		return fmt.Errorf("expected register to be %d for %s, got %d", expectedCmp.sreg, tname, cmp.sreg)
	}
	if cmp.cop != expectedCmp.cop {
		return fmt.Errorf("expected comparison operator to be %v for %s, got %v", expectedCmp.cop, tname, cmp.cop)
	}
	if !cmp.data.equal(expectedCmp.data) {
		return fmt.Errorf("expected data to be %v for %s, got %v", expectedCmp.data, tname, cmp.data)
	}
	return nil
}

// TestInterpretPayloadLoadOps tests interpretation of payload load operations.
// Most operations are direct output of nft binary commands. All stated commands
// should be preceded by nft --debug=netlink to generate matching operations.
func TestInterpretPayloadLoadOps(t *testing.T) {
	for _, test := range []interpretOperationTestAction{
		{
			tname:    "load bytes into verdict register",
			opStr:    "[ payload load 2b @ transport header + 0 => reg 0 ]",
			expected: nil,
		},
		// cmd: add rule ip tab ch tcp flags syn counter accept
		{
			tname:    "load 1 byte into 4-byte register",
			opStr:    "[ payload load 1b @ transport header + 13 => reg 9 ]",
			expected: mustCreatePayloadLoad(t, linux.NFT_PAYLOAD_TRANSPORT_HEADER, 13, 1, linux.NFT_REG32_01),
		},
		{
			tname:    "load 1 byte into 16-byte register",
			opStr:    "[ payload load 1b @ transport header + 13 => reg 1 ]",
			expected: mustCreatePayloadLoad(t, linux.NFT_PAYLOAD_TRANSPORT_HEADER, 13, 1, linux.NFT_REG_1),
		},
		// cmd: add rule ip tab ch tcp sport 80 counter accept
		{
			tname:    "load 2 bytes into 4-byte register no offset",
			opStr:    "[ payload load 2b @ transport header + 0 => reg 8 ]",
			expected: mustCreatePayloadLoad(t, linux.NFT_PAYLOAD_TRANSPORT_HEADER, 0, 2, linux.NFT_REG32_00),
		},
		{
			tname:    "load 2 bytes into 16-byte register no offset",
			opStr:    "[ payload load 2b @ transport header + 0 => reg 1 ]",
			expected: mustCreatePayloadLoad(t, linux.NFT_PAYLOAD_TRANSPORT_HEADER, 0, 2, linux.NFT_REG_1),
		},
		// cmd: add rule ip tab ch tcp dport 12345 counter accept
		{
			tname:    "load 2 bytes into 4-byte register with offset",
			opStr:    "[ payload load 2b @ transport header + 2 => reg 9 ]",
			expected: mustCreatePayloadLoad(t, linux.NFT_PAYLOAD_TRANSPORT_HEADER, 2, 2, linux.NFT_REG32_01),
		},
		{
			tname:    "load 2 bytes into 16-byte register with offset",
			opStr:    "[ payload load 2b @ transport header + 2 => reg 1 ]",
			expected: mustCreatePayloadLoad(t, linux.NFT_PAYLOAD_TRANSPORT_HEADER, 2, 2, linux.NFT_REG_1),
		},
		// cmd: add rule ip tab ch @th,24,24 0xabcdef counter accept
		{
			tname:    "load 3 bytes into 4-byte register",
			opStr:    "[ payload load 3b @ transport header + 3 => reg 10 ]",
			expected: mustCreatePayloadLoad(t, linux.NFT_PAYLOAD_TRANSPORT_HEADER, 3, 3, linux.NFT_REG32_02),
		},
		{
			tname:    "load 3 bytes into 16-byte register",
			opStr:    "[ payload load 3b @ transport header + 3 => reg 2 ]",
			expected: mustCreatePayloadLoad(t, linux.NFT_PAYLOAD_TRANSPORT_HEADER, 3, 3, linux.NFT_REG_2),
		},
		// cmd: add rule ip tab ch ip daddr 192.168.1.1 counter accept
		{
			tname:    "load 4 bytes into 4-byte register",
			opStr:    "[ payload load 4b @ network header + 16 => reg 12 ]",
			expected: mustCreatePayloadLoad(t, linux.NFT_PAYLOAD_NETWORK_HEADER, 16, 4, linux.NFT_REG32_04),
		},
		{
			tname:    "load 4 bytes into 16-byte register",
			opStr:    "[ payload load 4b @ network header + 16 => reg 1 ]",
			expected: mustCreatePayloadLoad(t, linux.NFT_PAYLOAD_NETWORK_HEADER, 16, 4, linux.NFT_REG_1),
		},
		// cmd: add rule ip tab ch ether saddr 01:23:45:67:89:ab counter drop
		{
			tname:    "load 6 bytes into 4-byte register",
			opStr:    "[ payload load 6b @ link header + 6 => reg 13 ]",
			expected: nil,
		},
		{
			tname:    "load 6 bytes into 16-byte register",
			opStr:    "[ payload load 6b @ link header + 6 => reg 3 ]",
			expected: mustCreatePayloadLoad(t, linux.NFT_PAYLOAD_LL_HEADER, 6, 6, linux.NFT_REG_3),
		},
		// cmd: add rule ip6 tab ch ip6 saddr 2001:db8::2 counter accept
		{
			tname:    "load 16 bytes into 4-byte register",
			opStr:    "[ payload load 16b @ network header + 8 => reg 10 ]",
			expected: nil,
		},
		{
			tname:    "load 16 bytes into 16-byte register",
			opStr:    "[ payload load 16b @ network header + 8 => reg 1 ]",
			expected: mustCreatePayloadLoad(t, linux.NFT_PAYLOAD_NETWORK_HEADER, 8, 16, linux.NFT_REG_1),
		},
		{
			tname:    "load >16 bytes into 16-byte register",
			opStr:    "[ payload load 20b @ network header + 16 => reg 1 ]",
			expected: nil,
		},
	} {
		t.Run(test.tname, func(t *testing.T) { checkOp(t, test, checkPayloadLoadOp) })
	}
}

// checkPayloadLoadOp checks that the given operation is a payload load
// operation and that it matches the expected payload load operation.
func checkPayloadLoadOp(tname string, expected operation, actual operation) error {
	expectedPdLoad := expected.(*payloadLoad)
	pdload, ok := actual.(*payloadLoad)
	if !ok {
		return fmt.Errorf("expected operation type to be PayloadLoad for %s, got %T", tname, actual)
	}
	if pdload.base != expectedPdLoad.base {
		return fmt.Errorf("expected payload base to be %v for %s, got %v", expectedPdLoad.base, tname, pdload.base)
	}
	if pdload.offset != expectedPdLoad.offset {
		return fmt.Errorf("expected offset to be %d for %s, got %d", expectedPdLoad.offset, tname, pdload.offset)
	}
	if pdload.blen != expectedPdLoad.blen {
		return fmt.Errorf("expected length to be %d for %s, got %d", expectedPdLoad.blen, tname, pdload.blen)
	}
	if pdload.dreg != expectedPdLoad.dreg {
		return fmt.Errorf("expected destination register to be %d for %s, got %d", expectedPdLoad.dreg, tname, pdload.dreg)
	}
	return nil
}

// TestInterpretPayloadSetOps tests interpretation of payload set operations.
// Most operations are direct output of nft binary commands. All stated commands
// should be preceded by nft --debug=netlink to generate matching operations.
func TestInterpretPayloadSetOps(t *testing.T) {
	for _, test := range []interpretOperationTestAction{
		// Simple checksum type tests.
		{
			tname:    "set checksum type, none",
			opStr:    "[ payload write reg 1 => 6b @ link header + 0 csum_type 0 csum_off 0 csum_flags 0x0 ]",
			expected: mustCreatePayloadSet(t, linux.NFT_PAYLOAD_LL_HEADER, 0, 6, linux.NFT_REG_1, linux.NFT_PAYLOAD_CSUM_NONE, 0, 0x0),
		},
		{
			tname:    "set checksum type, inet",
			opStr:    "[ payload write reg 1 => 6b @ link header + 0 csum_type 1 csum_off 0 csum_flags 0x0 ]",
			expected: mustCreatePayloadSet(t, linux.NFT_PAYLOAD_LL_HEADER, 0, 6, linux.NFT_REG_1, linux.NFT_PAYLOAD_CSUM_INET, 0, 0x0),
		},
		{
			tname:    "set checksum type, sctp", // not supported
			opStr:    "[ payload write reg 1 => 6b @ link header + 0 csum_type 2 csum_off 0 csum_flags 0x0 ]",
			expected: nil,
		},
		{
			tname:    "set out of range checksum type",
			opStr:    "[ payload write reg 1 => 6b @ link header + 0 csum_type 3 csum_off 0 csum_flags 0x0 ]",
			expected: nil,
		},
		// Simple checksum offset tests.
		{
			tname:    "set valid offset",
			opStr:    "[ payload write reg 1 => 6b @ link header + 0 csum_type 0 csum_off 100 csum_flags 0x0 ]",
			expected: mustCreatePayloadSet(t, linux.NFT_PAYLOAD_LL_HEADER, 0, 6, linux.NFT_REG_1, linux.NFT_PAYLOAD_CSUM_NONE, 100, 0x0),
		},
		{
			tname:    "set negative checksum offset",
			opStr:    "[ payload write reg 1 => 6b @ link header + 100 csum_type 1 csum_off -1 csum_flags 0x0 ]",
			expected: nil,
		},
		// Simple checksum flags tests.
		{
			tname:    "set checksum flags, L4 with psuedoheader flag",
			opStr:    "[ payload write reg 1 => 6b @ link header + 0 csum_type 0 csum_off 0 csum_flags 0x1 ]",
			expected: mustCreatePayloadSet(t, linux.NFT_PAYLOAD_LL_HEADER, 0, 6, linux.NFT_REG_1, linux.NFT_PAYLOAD_CSUM_NONE, 0, linux.NFT_PAYLOAD_L4CSUM_PSEUDOHDR),
		},
		{
			tname:    "set invalid checksum flags",
			opStr:    "[ payload write reg 1 => 6b @ link header + 0 csum_type 0 csum_off 0 csum_flags 0x2 ]",
			expected: nil,
		},
		// Invalid register tests.
		{
			tname:    "set from verdict register",
			opStr:    "[ payload write reg 0 => 4b @ link header + 0 csum_type 0 csum_off 0 csum_flags 0x0 ]",
			expected: nil,
		},
		{
			tname:    "set >4 bytes from 4-byte register",
			opStr:    "[ payload write reg 9 => 6b @ link header + 0 csum_type 0 csum_off 0 csum_flags 0x0 ]",
			expected: nil,
		},
		{
			tname:    "set >16 bytes from 16-byte register",
			opStr:    "[ payload write reg 2 => 20b @ link header + 0 csum_type 0 csum_off 0 csum_flags 0x0 ]",
			expected: nil,
		},

		// Valid tests.
		// Note: It doesn't seem like the nft binary ever outputs payload set ops
		// that have an odd offset or length and checksumming on. This makes sense
		// because the offset and length are specified in bytes, but the checksum is
		// calculated in half-words (2-bytes), which means the checksum calculation
		// is only valid if the offset and length are even. However, the linux
		// kernel does not specifically enforce this, so on linux it's technically
		// possible to declare payload set operations that undoubtedly result in
		// invalid checksums. Since the nft binary is what generates our input, we
		// do not test these edge cases either.

		// cmd: add rule ip tab ch @nh,24,8 set 0xab
		{
			tname:    "set 1 byte from 4-byte register with csum NONE and no flags",
			opStr:    "[ payload write reg 8 => 1b @ network header + 3 csum_type 0 csum_off 0 csum_flags 0x0 ]",
			expected: mustCreatePayloadSet(t, linux.NFT_PAYLOAD_NETWORK_HEADER, 3, 1, linux.NFT_REG32_00, linux.NFT_PAYLOAD_CSUM_NONE, 0, 0x0),
		},
		{
			tname:    "set 1 byte from 16-byte register with csum NONE and no flags",
			opStr:    "[ payload write reg 1 => 1b @ network header + 4 csum_type 0 csum_off 0 csum_flags 0x0 ]",
			expected: mustCreatePayloadSet(t, linux.NFT_PAYLOAD_NETWORK_HEADER, 4, 1, linux.NFT_REG_1, linux.NFT_PAYLOAD_CSUM_NONE, 0, 0x0),
		},
		// cmd: add rule ip tab ch tcp sport set 80
		{
			tname:    "set 2 bytes from 4-byte register with csum INET and no flags",
			opStr:    "[ payload write reg 9 => 2b @ transport header + 0 csum_type 1 csum_off 16 csum_flags 0x0 ]",
			expected: mustCreatePayloadSet(t, linux.NFT_PAYLOAD_TRANSPORT_HEADER, 0, 2, linux.NFT_REG32_01, linux.NFT_PAYLOAD_CSUM_INET, 16, 0x0),
		},
		{
			tname:    "set 2 bytes from 16-byte register with csum INET and no flags",
			opStr:    "[ payload write reg 2 => 2b @ transport header + 0 csum_type 1 csum_off 16 csum_flags 0x0 ]",
			expected: mustCreatePayloadSet(t, linux.NFT_PAYLOAD_TRANSPORT_HEADER, 0, 2, linux.NFT_REG_2, linux.NFT_PAYLOAD_CSUM_INET, 16, 0x0),
		},
		// cmd: add rule ip tab ch @ll,24,24 set 0xabcdef
		{
			tname:    "set 3 bytes from 4-byte register with csum NONE and no flags",
			opStr:    "[ payload write reg 10 => 3b @ link header + 3 csum_type 0 csum_off 0 csum_flags 0x0 ]",
			expected: mustCreatePayloadSet(t, linux.NFT_PAYLOAD_LL_HEADER, 3, 3, linux.NFT_REG32_02, linux.NFT_PAYLOAD_CSUM_NONE, 0, 0x0),
		},
		{
			tname:    "set 3 bytes from 16-byte register with csum NONE and no flags",
			opStr:    "[ payload write reg 2 => 3b @ link header + 3 csum_type 0 csum_off 0 csum_flags 0x0 ]",
			expected: mustCreatePayloadSet(t, linux.NFT_PAYLOAD_LL_HEADER, 3, 3, linux.NFT_REG_2, linux.NFT_PAYLOAD_CSUM_NONE, 0, 0x0),
		},
		// cmd: add rule ip tab ch ip daddr set 192.168.1.1
		{
			tname:    "set 4 bytes from 4-byte register with csum INET and pseudoheader flag",
			opStr:    "[ payload write reg 11 => 4b @ network header + 16 csum_type 1 csum_off 10 csum_flags 0x1 ]",
			expected: mustCreatePayloadSet(t, linux.NFT_PAYLOAD_NETWORK_HEADER, 16, 4, linux.NFT_REG32_03, linux.NFT_PAYLOAD_CSUM_INET, 10, linux.NFT_PAYLOAD_L4CSUM_PSEUDOHDR),
		},
		{
			tname:    "set 4 bytes from 4-byte register with csum INET and pseudoheader flag",
			opStr:    "[ payload write reg 3 => 4b @ network header + 16 csum_type 1 csum_off 10 csum_flags 0x1 ]",
			expected: mustCreatePayloadSet(t, linux.NFT_PAYLOAD_NETWORK_HEADER, 16, 4, linux.NFT_REG_3, linux.NFT_PAYLOAD_CSUM_INET, 10, linux.NFT_PAYLOAD_L4CSUM_PSEUDOHDR),
		},
		// cmd: add rule ip tab ch ether saddr set 01:23:45:67:89:ab
		{
			tname:    "set 6 bytes from 16-byte register with csum NONE and no flags",
			opStr:    "[ payload write reg 4 => 6b @ link header + 6 csum_type 0 csum_off 0 csum_flags 0x0 ]",
			expected: mustCreatePayloadSet(t, linux.NFT_PAYLOAD_LL_HEADER, 6, 6, linux.NFT_REG_4, linux.NFT_PAYLOAD_CSUM_NONE, 0, 0x0),
		},
		// cmd: add rule ip6 tab ch ip6 saddr set 2001:db8::2
		{
			tname:    "set 16 bytes from 16-byte register with csum NONE and psuedoheader flag",
			opStr:    "[ payload write reg 1 => 16b @ network header + 8 csum_type 0 csum_off 0 csum_flags 0x1 ]",
			expected: mustCreatePayloadSet(t, linux.NFT_PAYLOAD_NETWORK_HEADER, 8, 16, linux.NFT_REG_1, linux.NFT_PAYLOAD_CSUM_NONE, 0, linux.NFT_PAYLOAD_L4CSUM_PSEUDOHDR),
		},
	} {
		t.Run(test.tname, func(t *testing.T) { checkOp(t, test, checkPayloadSetOp) })
	}
}

// checkPayloadSetOp checks that the given operation is a payload set
// operation and that it matches the expected payload set operation.
func checkPayloadSetOp(tname string, expected operation, actual operation) error {
	expectedPdSet := expected.(*payloadSet)
	pdset, ok := actual.(*payloadSet)
	if !ok {
		return fmt.Errorf("expected operation type to be PayloadLoad for %s, got %T", tname, actual)
	}
	if pdset.base != expectedPdSet.base {
		return fmt.Errorf("expected payload base to be %v for %s, got %v", expectedPdSet.base, tname, pdset.base)
	}
	if pdset.offset != expectedPdSet.offset {
		return fmt.Errorf("expected offset to be %d for %s, got %d", expectedPdSet.offset, tname, pdset.offset)
	}
	if pdset.blen != expectedPdSet.blen {
		return fmt.Errorf("expected length to be %d for %s, got %d", expectedPdSet.blen, tname, pdset.blen)
	}
	if pdset.sreg != expectedPdSet.sreg {
		return fmt.Errorf("expected destination register to be %d for %s, got %d", expectedPdSet.sreg, tname, pdset.sreg)
	}
	if pdset.csumType != expectedPdSet.csumType {
		return fmt.Errorf("expected checksum type to be %d for %s, got %d", expectedPdSet.csumType, tname, pdset.csumType)
	}
	if pdset.csumOffset != expectedPdSet.csumOffset {
		return fmt.Errorf("expected checksum offset to be %d for %s, got %d", expectedPdSet.csumOffset, tname, pdset.csumOffset)
	}
	if pdset.csumFlags != expectedPdSet.csumFlags {
		return fmt.Errorf("expected checksum flags to be %b for %s, got %b", expectedPdSet.csumFlags, tname, pdset.csumFlags)
	}
	return nil
}

// TestInterpretRule tests the interpretation of basic and general rules as a
// list of operations.
func TestInterpretRule(t *testing.T) {
	for _, test := range []struct {
		tname    string
		ruleStr  string
		expected *Rule // will be nil if an error is expected
	}{
		{
			tname:    "empty ruleset",
			ruleStr:  ``,
			expected: &Rule{},
		},
		{
			tname: "empty ruleset with excess whitespace",
			ruleStr: `		


			`,
			expected: &Rule{},
		},
	} {
		t.Run(test.tname, func(t *testing.T) {
			rule, err := InterpretRule(test.ruleStr)
			if test.expected == nil {
				if err == nil {
					t.Fatalf("unexpected interpretation success for %s", test.tname)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected interpretation error for %s: %v", test.tname, err)
			}

			if len(rule.ops) != len(test.expected.ops) {
				t.Fatalf("expected %d operations for %s, got %d", len(test.expected.ops), test.tname, len(rule.ops))
			}

			// Checks each operation in the rule with the appropriate check function.
			for i, op := range rule.ops {
				testOp := test.expected.ops[i]
				switch testOp.(type) {
				case *immediate:
					if err := checkImmediateOp(test.tname, testOp, op); err != nil {
						t.Fatalf(err.Error())
					}
				case *comparison:
					if err := checkComparisonOp(test.tname, testOp, op); err != nil {
						t.Fatalf(err.Error())
					}
				// TODO(b/345684870): cases will be added here as more types are supported.
				default:
					t.Fatalf("unexpected operation type for %s: %T", test.tname, testOp)
				}
			}
		})
	}
}
