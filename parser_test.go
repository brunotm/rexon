package rexon

import (
	"context"
	"testing"
)

var (
	dataLine = []byte(
		`2       0 fd0 1090 0 8720 46276 0 0 0 0 0 46276 46276
		8       0 sda 5154769 15140912 164152460 3212508 1476128 29608370 261479164 76894036 0 2670448 80258772
		8       1 sda1 193 84 10754 460 5 1 12 0 0 436 460
		8       2 sda2 4 0 8 8 0 0 0 0 0 8 8
		8       5 sda5 5154478 15140828 164137442 3211724 1476123 29608369 261479152 76894036 0 2670244 80256780
		8      16 sdb 244976381 183541 2231075482 1210800896 545032062 438733098 19675187544 2307970752 0 412820256 3613103028
		8      17 sdb1 244976342 183541 2231073282 1210800836 545032062 438733098 19675187544 2307970752 0 412820068 3612915632
		8      32 sdc 2837 66 302104 13256 119848 18128 35620216 504928 0 19180 518136
	  253       0 dm-0 35230 0 2325842 197552 2125166 0 64052832 1354520 0 324924 1552564
	  253       1 dm-1 20262834 0 162111560 9458536 29130817 0 233046536 4282656576 0 2377172 94596
	  253       2 dm-2 62515 0 2653624 129068 26815219 0 262524640 87229100 0 18302076 87441980
	  253       3 dm-3 245103866 0 2228416474 1214164760 955225124 0 19412662904 2534810408 0 405566856 3757302528

	  `)

	dataMLine = []byte(`
			message aaammmkkklll
id 8879789.9
vmm 7hgj cdd xxkkll

message bbmm
id 67
vmm bcn cdd llmm`)
)

func TestParserLineLine(t *testing.T) {

	values := []*Value{
		MustNewValue("maj", Number),
		MustNewValue("min", Number),
		MustNewValue("device", String)}

	p, err := NewParser(values, LineRegex(`(\d+)\s+(\d+)\s+(.*?)\s+`))
	if err != nil {
		t.Fatal(err)
	}

	for d := range p.ParseBytes(context.Background(), dataLine) {
		if d.Errors != nil {
			t.Fatal(d.Errors)
		}
		t.Logf("%#v\n", d)
	}

}

func TestParserMLineLine(t *testing.T) {
	values := []*Value{
		MustNewValue("message", String),
		MustNewValue("id", Number),
		MustNewValue("vmm", String),
		MustNewValue("cdd", String)}

	p, err := NewParser(
		values,
		LineRegex(`(?m)\s*message\s*(\w+)\nid\s*([-+]?[0-9]*\.?[0-9]+)\nvmm\s*(\w+)\s*cdd\s*(\w+)`))

	if err != nil {
		t.Fatal(err)
	}
	for d := range p.ParseBytes(context.Background(), dataMLine) {
		if d.Errors != nil {
			t.Fatal(d.Errors)
		}
		t.Logf("%#v\n", d)
	}

}

func TestParserMLineSet(t *testing.T) {
	values := []*Value{
		MustNewValue("message", String, ValueRegex(`\s*message\s*(\w+)\s*`)),
		MustNewValue("id", Number, Round(2), ValueRegex(`id\s*([-+]?[0-9]*\.?[0-9]+)\s*`)),
		MustNewValue("vmm", String, ValueRegex(`vmm\s*(\w+)\s*`)),
		MustNewValue("cdd", String, ValueRegex(`cdd\s*(\w+)`))}

	p, err := NewParser(values, StartTag(`message.*`))
	if err != nil {
		t.Fatal(err)
	}
	for d := range p.ParseBytes(context.Background(), dataMLine) {
		if d.Errors != nil {
			t.Fatal(d.Errors)
		}
		t.Logf("%#v\n", d)
	}

}

func BenchmarkLineLine(b *testing.B) {
	values := []*Value{
		MustNewValue("maj", Number),
		MustNewValue("min", Number),
		MustNewValue("device", String)}

	p, err := NewParser(values, LineRegex(`(\d+)\s+(\d+)\s+(.*?)\s+`))

	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		for range p.ParseBytes(context.Background(), dataMLine) {
		}
	}
}

func BenchmarkLineSet(b *testing.B) {
	values := []*Value{
		MustNewValue("maj", Number, Round(2), ValueRegex(`(\d+)\s+\d+\s+.*?\s+`)),
		MustNewValue("min", Number, Round(2), ValueRegex(`\d+\s+(\d+)\s+.*?\s+`)),
		MustNewValue("device", String, ValueRegex(`\d+\s+\d+\s+(.*?)\s+`))}

	p, err := NewParser(values)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		for range p.ParseBytes(context.Background(), dataMLine) {
		}
	}
}

func BenchmarkMLineLine(b *testing.B) {
	values := []*Value{
		MustNewValue("message", String),
		MustNewValue("id", Number),
		MustNewValue("vmm", String),
		MustNewValue("cdd", String)}

	p, err := NewParser(
		values,
		LineRegex(`(?m)\s*message\s*(\w+)\nid\s*([-+]?[0-9]*\.?[0-9]+)\nvmm\s*(\w+)\s*cdd\s*(\w+)`))

	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		for range p.ParseBytes(context.Background(), dataMLine) {
		}
	}
}

func BenchmarkMLineSet(b *testing.B) {
	values := []*Value{
		MustNewValue("message", String, ValueRegex(`\s*message\s*(\w+)\s*`)),
		MustNewValue("id", Number, Round(2), ValueRegex(`id\s*([-+]?[0-9]*\.?[0-9]+)\s*`)),
		MustNewValue("vmm", String, ValueRegex(`vmm\s*(\w+)\s*`)),
		MustNewValue("cdd", String, ValueRegex(`cdd\s*(\w+)`))}

	p, err := NewParser(values, StartTag(`message.*`))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		for range p.ParseBytes(context.Background(), dataMLine) {
		}
	}
}
