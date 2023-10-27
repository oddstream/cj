# util.tcl

proc union { lista listb } {
	set result $lista

	foreach elem $listb {
		if { [lsearch -exact $lista $elem] == -1 } {
			lappend result $elem
		}
	}
	return $result
}

proc intersection { lista listb } {
	set result {}

	foreach elem $listb {
		if { [lsearch -exact $lista $elem] != -1 } {
			lappend result $elem
		}
	}
	return $result
}

proc exclusion { lista listb } {
	set result {}

	foreach elem $lista {
		if { [lsearch -exact $listb $elem] == -1 } {
			lappend result $elem
		}
	}
	return $result
}