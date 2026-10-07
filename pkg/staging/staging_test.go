package staging

import (
	"regexp"
	"sync"
	"testing"
)

var stagingTableName = regexp.MustCompile(`^__omni_stage_[0-9a-f]{32}_[a-z0-9_]*_[0-9a-f]{32}$`)

func TestScopeIsolatesQueriesAndReusesNames(t *testing.T) {
	first, err := NewScope()
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewScope()
	if err != nil {
		t.Fatal(err)
	}

	firstName, err := first.TableName("aws.ec2.instances")
	if err != nil {
		t.Fatal(err)
	}
	reusedName, err := first.TableName("aws.ec2.instances")
	if err != nil {
		t.Fatal(err)
	}
	secondName, err := second.TableName("aws.ec2.instances")
	if err != nil {
		t.Fatal(err)
	}

	if firstName != reusedName {
		t.Fatalf("same logical table got different names: %q != %q", firstName, reusedName)
	}
	if firstName == secondName {
		t.Fatalf("different query scopes share a table name: %q", firstName)
	}
	if !stagingTableName.MatchString(firstName) {
		t.Fatalf("table name is not a safe staging identifier: %q", firstName)
	}
	if got := first.Names(); len(got) != 1 || got[0] != firstName {
		t.Fatalf("unexpected allocated names: %v", got)
	}
}

func TestScopeDistinguishesSanitizedNameCollisions(t *testing.T) {
	scope, err := NewScope()
	if err != nil {
		t.Fatal(err)
	}

	first, err := scope.TableName("a-b")
	if err != nil {
		t.Fatal(err)
	}
	second, err := scope.TableName("a_b")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("distinct logical tables share a physical name: %q", first)
	}
}

func TestScopeRejectsEmptyName(t *testing.T) {
	scope, err := NewScope()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := scope.TableName(""); err == nil {
		t.Fatal("expected empty logical name to fail")
	}
}

func TestScopeAllocatesConcurrently(t *testing.T) {
	scope, err := NewScope()
	if err != nil {
		t.Fatal(err)
	}

	const calls = 32
	names := make([]string, calls)
	var group sync.WaitGroup
	group.Add(calls)
	for index := range calls {
		go func() {
			defer group.Done()
			name, nameErr := scope.TableName("google.compute.instances")
			if nameErr != nil {
				t.Errorf("allocate table name: %v", nameErr)
				return
			}
			names[index] = name
		}()
	}
	group.Wait()

	for _, name := range names {
		if name != names[0] {
			t.Fatalf("concurrent allocations returned different names: %v", names)
		}
	}
	if got := scope.Names(); len(got) != 1 {
		t.Fatalf("got %d allocated names, want 1: %v", len(got), got)
	}
}
