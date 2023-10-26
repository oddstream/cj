# commonplace journal

puts "Tcl $tcl_patchLevel"		;# 8.6.12
puts "Tk [package require Tk]"	;# 8.6.12

set defaultFont {"JetBrains Mono" 12}
set monthNames {"---" "January" "February" "March" "April" "May" "June" "July" "August" "September" "October" "November" "December"}
set monthAbbrs {"---" "Jan" "Feb" "Mar" "Apr" "May" "Jun" "Jul" "Aug" "Sep" "Oct" "Nov" "Dec"}

# could store the two importants dates (today and displayed)
# as separate variables, an array or a dict

# https://tcl.tk/man/tcl8.6/TclCmd/clock.htm
set todayYear [clock format [clock seconds] -format "%Y"]		;# four-digit calendar year

set todayMonth [clock format [clock seconds] -format "%m"]		;# the number of the month (01-12) with exactly two digits
set todayMonth [string trimleft $todayMonth 0]					;# strip any leading 0

set todayDay [clock format [clock seconds] -format "%d"]		;# the number of the day of the month, as two decimal digits
set todayDay [string trimleft $todayDay 0]						;# strip any leading 0

set displayedYear $todayYear
set displayedMonth $todayMonth
set displayedDay $todayDay

puts [format "%d-%d-%d" $displayedYear $displayedMonth $displayedDay]
puts [format "%s-%s-%s" $displayedYear $displayedMonth $displayedDay]

set searchString ""

# $ ncal -bh1
#     October 2023
# Mo Tu We Th Fr Sa Su
#                    1
#  2  3  4  5  6  7  8
#  9 10 11 12 13 14 15
# 16 17 18 19 20 21 22
# 23 24 25 26 27 28 29
# 30 31

# -m month
# -y year
# -1 display one month (put this at the end)
# to get April 1962, use ncal -bh -m April -y 1962 -1
#      April 1962
# Mo Tu We Th Fr Sa Su
#                    1
#  2  3  4  5  6  7  8
#  9 10 11 12 13 14 15
# 16 17 18 19 20 21 22
# 23 24 25 26 27 28 29
# 30

# date and filename string mangling

proc monthNameToNumber {m} {
	if {![string is integer -strict $m]} { puts "ERROR $m is not an integer" }
	global monthNames
	set i [lsearch $monthNames $m]
	return $i
}

proc monthName {m} {
	global monthNames
	return [lindex $monthNames $m]
}

proc filename {y m d} {
	if {![string is integer -strict $y]} { puts "ERROR $y is not an integer" }
	if {![string is integer -strict $m]} { puts "ERROR $m is not an integer" }
	if {![string is integer -strict $d]} { puts "ERROR $d is not an integer" }
	# day, month are used and displayed internally as 1 .. 31 but needs to be 01 .. 31 for filename
	return [file join / home gilbert .cj Default $y [format "%02d" $m] "[format "%02d" $d]\.txt"]
}

proc displayedFilename {} {
	# nb the filename isn't displayed, but is made from the displayed* global variables
	global displayedYear displayedMonth displayedDay
	return [filename $displayedYear $displayedMonth $displayedDay]
}

proc directory {} {
	return [file join / home gilbert .cj Default]
}

proc filenameToYMD {fname} {
	set fname [string range $fname [expr [string length [directory]] + 1] end]
	set fname [regexp -all -inline "\[0-9\]+/\[0-9\]+/\[0-9\]+" $fname]
	set ymd [split $fname /]
	lassign $ymd y m d
	set m [string trimleft $m 0]
	set d [string trimleft $d 0]
	return "$y $m $d" ;# use lassign to get these
}

proc filenameToISO8061 {fname} {
	set date [filenameToYMD $fname]
	lassign $date y m d
	return [format "%s-%s-%s" $y $m $d]
}

proc ISO8061ToDate {txt} {
	return [split $txt "-"]
}

proc ISO8061ToFilename {txt} {
	set date [ISO8061ToDate $txt]
	lassign $date y m d
	return [file join [directory] $y [format "%02d" $m] "[format "%02d" $d]\.txt"]
}

# note

proc isNoteDirty? {} {
	return [.pw.note edit modified]
}

proc getNoteText {} {
	# end goes beyond the end of the text, adding a new line, so subtract one character
	set txt [.pw.note get 1.0 "end - 1 c"]
}

proc setNoteText {txt} {
	.pw.note configure -undo false
	.pw.note delete 1.0 end
	.pw.note insert end $txt
	.pw.note configure -undo true
	.pw.note edit modified false
	focus .pw.note
}

proc loadDisplayedNote {} {
	if {[file exists [displayedFilename]]} {
		set f [open [displayedFilename] r]
		set fileContents [read $f]
		close $f
		setNoteText $fileContents
	} else {
		setNoteText ""
	}
	setWindowTitle
}

proc saveDisplayedNote {} {
	if {![isNoteDirty?]} {
		puts "displayed note is not dirty"
		return
	}
	set txt [getNoteText]
	set fname [displayedFilename]

	if { [string length $txt] == 0 && [file exists $fname] } {
		file delete $fname
		puts "deleted $fname"
	} else {
		set f [open $fname w+]	;# Open the file for reading and writing. Truncate it if it exists. If it does not exist, create a new file.
		puts $f $txt
		close $f
		puts "saved $fname"
	}
}

# search

proc doSearch {} {
	global searchString
	if { $searchString eq "" } {
		return
	}

	.pw.left.findframe.found delete 0 end

	set output ""
	set grepstatus 0
	try {
		# -I exclude binary files
		set output [exec grep --fixed-strings --recursive --ignore-case --files-with-matches --exclude-dir='.*' -I $searchString [directory]]
	} trap CHILDSTATUS {results options} {
    	set grepstatus [lindex [dict get $options -errorcode] 2]
		# puts [string cat "grepstatus " $grepstatus " results " $results]
		# status 0 line selected
		# status 1 no line selected result "child process exited abnormally"
		# status 2 error
	}
	if { $grepstatus != 0 } {
		return
	}
	set foundLines [split $output "\n"]
	foreach line $foundLines {
		.pw.left.findframe.found insert end [filenameToISO8061 $line]
		# puts $line
	}
}

# calendar

proc decorateDayButtons {} {
	proc decorate {targetday color} {
		# iterate through children of .pw.left.calframe
		# looking for children that end in a number
		foreach button $[winfo children .pw.left.calframe] {
			set day [regexp -inline "\[0-9\]+$" $button]
			if { [string length $day] == 0 } {
				continue
			}
			if { $targetday == "" } {
				$button configure -background $color
			} elseif { $day == $targetday } {
				$button configure -background $color
			}
		}
	}
	global displayedDay displayedMonth displayedYear
	global todayDay todayMonth todayYear

	# clear all old decorations
	set defaultcolor [.pw.left.calframe.monthYear cget -activebackground]
	decorate "" $defaultcolor

	# decorate today, if it's being shown
	if { $displayedMonth == $todayMonth && $displayedYear == $todayYear } {
		decorate $todayDay #FF0
	}

	# decorate currently selected day
	foreach button $[winfo children .pw.left.calframe] {
		decorate $displayedDay #080
	}
}

proc createRowOfDayButtons {row line} {
	# puts [format "%d %s" $row $line]
	set days [regexp -all -inline {\S+} $line]
	set padding [expr 7 - [llength $days]]
	set col 0
	if {$padding > 0} {
		set first [lindex $days 0]
		if {$first eq "1"} {
			set col $padding
		}
	}
	foreach day $days {
		button .pw.left.calframe.b$day -text $day -command "setDay $day"
		grid .pw.left.calframe.b$day -row $row -column $col -sticky news
		incr col
	}
}

proc clearDayButtons {} {
	# doing this creates flicker when only changing day
	for {set i 1} {$i < 32} {incr i} {
		# destroy creates no error if window does not exist
		destroy .pw.left.calframe.b$i
	}
}

proc createDayButtons {calendarLines} {
	set row 2
	for {set i 2} {$i < [llength $calendarLines]} {incr i} {
		createRowOfDayButtons $row [lindex $calendarLines $i]
		incr row
	}
	decorateDayButtons
}

proc refreshCalendar {} {
	global monthNames displayedDay displayedMonth displayedYear
	# puts [string cat "refreshCalendar " $displayedYear " " $displayedMonth " " $displayedDay]

	clearDayButtons
	set calendarLines [split [exec ncal -bh -m $displayedMonth -y $displayedYear -1] "\n"]
	# day names do not change
	# set dayNames [regexp -all -inline {\S+} [lindex $calendarLines 1]]
	.pw.left.calframe.monthYear configure -text [string cat [monthName $displayedMonth] " " $displayedYear]
	createDayButtons $calendarLines
}

proc prevMonth {} {
	global displayedDay displayedMonth displayedYear
	if {$displayedMonth == 1} {
		set displayedMonth 12
		incr displayedYear -1
	} else {
		incr displayedMonth -1
	}
	setDay 1
	refreshCalendar
}

proc nextMonth {} {
	global displayedDay displayedMonth displayedYear
	if {$displayedMonth == 12} {
		set displayedMonth 1
		incr displayedYear
	} else {
		incr displayedMonth
	}
	setDay 1
	refreshCalendar
}

proc setDay {day} {
	global displayedDay displayedMonth displayedYear
	set displayedDay $day
	decorateDayButtons
	loadDisplayedNote
}

proc setYearMonthDay {y m d} {
	global displayedDay displayedMonth displayedYear
	if { $y == $displayedYear && $m == $displayedMonth } {
		setDay $d
	} else {
		set displayedYear $y
		set displayedMonth $m
		set displayedDay $d
		refreshCalendar
		loadDisplayedNote
	}
}

# ui

proc setWindowTitle {} {
	global displayedDay displayedMonth displayedYear
	wm title . [format "%s %s %s" $displayedDay [monthName $displayedMonth] $displayedYear]
}

# -b oldstyle format for ncal output
# -h turn off highlighting of current date
# -1 only show one month, has to be last
#puts "ncal -bh -m $displayedMonth -y $displayedYear -1"
set calendarLines [split [exec ncal -bh -m $displayedMonth -y $displayedYear -1] "\n"]
#foreach line $calendarLines {
#	puts $line
#}

#set monthYear [regexp -all -inline {\S+} [lindex $calendarLines 0]]
#set displayedMonth [lindex $monthYear 0]
#set displayedYear [lindex $monthYear 1]
set dayNames [regexp -all -inline {\S+} [lindex $calendarLines 1]]
#puts $dayNames

panedwindow .pw -orient horizontal -showhandle 1

# outer frame for all widgets on left hand side of panedwindow
ttk::labelframe .pw.left -text "Left Frame"

# create a 7-column frame to hold the header, day names and day buttons
ttk::labelframe .pw.left.calframe -text "Calendar Frame"
grid .pw.left.calframe -sticky news

# add "< October 2023 >" to first (0th) row
button .pw.left.calframe.prevMonth -text "<" -command prevMonth
grid .pw.left.calframe.prevMonth -row 0 -column 0
label .pw.left.calframe.monthYear -text [string cat [monthName $displayedMonth] " " $displayedYear]
grid .pw.left.calframe.monthYear -row 0 -column 1 -columnspan 5
button .pw.left.calframe.nextMonth -text ">" -command nextMonth
grid .pw.left.calframe.nextMonth -row 0 -column 6

# add "Mo Tu We Th Fr Sa Su" to second (1st) row
for {set i 0} {$i < 7} {incr i} {
	# window name cannot start with uppercase letter
	set Da [lindex $dayNames $i]
	set da [string tolower $Da]
	label .pw.left.calframe.$da -text $Da
	grid .pw.left.calframe.$da -row 1 -column $i -sticky news
}

createDayButtons $calendarLines

ttk::labelframe .pw.left.findframe -text "Find Frame"
grid .pw.left.findframe

# ttk::entry .pw.left.search
entry .pw.left.findframe.searchEntry -textvariable searchString
grid .pw.left.findframe.searchEntry -row 1 -column 1
grid columnconfigure .pw.left.findframe.searchEntry 1 -weight 5

button .pw.left.findframe.searchButton -text "Search" -command doSearch
grid .pw.left.findframe.searchButton -row 1 -column 2

listbox .pw.left.findframe.found -selectmode browse -font $defaultFont
bind .pw.left.findframe.found <<ListboxSelect>> {
	# puts -nonewline "listbox select "
	# puts -nonewline [%W curselection]
	# puts [.pw.left.findframe.found get [%W curselection]]
	if { [llength [%W curselection]] == 1 } {
		set i [lindex [%W curselection] 0]
		if { [string is integer $i] } {
			set txt [.pw.left.findframe.found get $i]
			set txt [ISO8061ToDate $txt]
			lassign $txt y m d
			setYearMonthDay $y $m $d
		}
	}
}
grid .pw.left.findframe.found -row 2 -column 1
grid columnconfigure .pw.left.findframe.found 1 -weight 5

# right hand side of panedwindow is just a text widget
text .pw.note -wrap word -undo 1 -font $defaultFont

.pw add .pw.left .pw.note
pack .pw -fill both -expand yes

loadDisplayedNote

# set foo [file join ~ .cj Default $displayedYear [monthNameToNumber $displayedMonth]]
# puts [format "%d %s" [file isdirectory $foo] $foo]
# puts [format "%d %s" [file exists [displayedFilename]] [displayedFilename]]

# https://www.pythontutorial.net/tkinter/tkinter-event-binding/
menu .m
.m add command -label "Cut" -command {puts "Cut"}
.m add command -label "Copy" -command {puts "Copy"}
.m add command -label "Paste" -command {puts "Paste"}
.m add command -label Save -underline 0 -command saveDisplayedNote -accelerator <Control-s>

bind .pw.note <Button-3> {tk_popup .m %X %Y}
bind . <Control-s> {.m invoke Save}

wm protocol . WM_DELETE_WINDOW {
	set response [tk_messageBox -type yesno -message "Really quit?"]
	if {$response eq "yes"} {
		# save displayed note
		destroy .
	}
}