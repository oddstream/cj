# commonplace journal

puts "Tcl $tcl_patchLevel"		;# 8.6.12
puts "Tk [package require Tk]"	;# 8.6.12

source util.tcl

set defaultFont {"JetBrains Mono" 12}
set todayColor #FF0
set activeColor #080
set monthNames {"---" "January" "February" "March" "April" "May" "June" "July" "August" "September" "October" "November" "December"}
set monthAbbrs {"---" "Jan" "Feb" "Mar" "Apr" "May" "Jun" "Jul" "Aug" "Sep" "Oct" "Nov" "Dec"}
set searchString ""

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

proc displayedDirectory {} {
	global displayedYear displayedMonth
	return [file join / home gilbert .cj Default $displayedYear [format "%02d" $displayedMonth]]
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
		set dir [displayedDirectory]
		if { ![file isdirectory $dir] } {
			puts "making $dir"
			file mkdir $dir
		}
		set f [open $fname w+]	;# Open the file for reading and writing. Truncate it if it exists. If it does not exist, create a new file.
		puts $f $txt
		close $f
		puts "saved $fname"
	}
}

# search

proc cancelDialogCommand {} {
	destroy .dialog
}

proc showFindDialog {title label command} {
	global searchString
	set searchString ""

	toplevel .dialog

	# hide .dialog while we build it
	wm withdraw .dialog

	ttk::entry .dialog.findText -textvariable searchString
	ttk::button .dialog.findButton -text $label -command $command
	ttk::button .dialog.cancelButton -text "Cancel" -command {cancelDialogCommand}

	grid config .dialog.findText -column 1 -row 0
	grid config .dialog.findButton -column 2 -row 0
	grid config .dialog.cancelButton -column 3 -row 0

	wm title .dialog $title
	wm protocol .dialog WM_DELETE_WINDOW {
		.dialog.cancelButton invoke
	}
	wm transient .dialog .

	bind .dialog <Return> {.dialog.findButton invoke}
	bind .dialog <Escape> {.dialog.cancelButton invoke}

	# ready to display dialog
	wm deiconify .dialog

	# make .dialog modal
	catch {tk visibility .dialog}
	focus .dialog.findText
	catch {grab set .dialog}
	catch {tkwait window .dialog}
}

proc findCommand {} {
	proc do {} {
		global searchString
		if { $searchString eq "" } {
			return
		}
		set foundLines [makeFoundList]
		if { [llength $foundLines] == 0 } {
			return
		}
		.pw.left.found delete 0 end
		foreach line $foundLines {
			.pw.left.found insert end [filenameToISO8061 $line]
		}
		cancelDialogCommand
	}
	showFindDialog "Find Notes" "Find" do
}

proc widenCommand {} {
	proc do {} {
		global searchString
		if { $searchString eq "" } {
			return
		}
		set foundLines [makeFoundList]
		if { [llength $foundLines] == 0 } {
			return
		}
		set oldlines [.pw.left.found get 0 end]
		set oldlines2 {}
		foreach line $oldlines {
			lappend oldlines2 [ISO8061ToFilename $line]
		}
		set newlines [union $oldlines2 $foundLines]
		.pw.left.found delete 0 end
		foreach line $newlines {
			.pw.left.found insert end [filenameToISO8061 $line]
		}
		cancelDialogCommand
	}
	if { [.pw.left.found size] == 0 } {
		return
	}
	showFindDialog "Find More Notes" "Widen" do
}

proc narrowCommand {} {
	proc do {} {
		global searchString
		if { $searchString eq "" } {
			return
		}
		set foundLines [makeFoundList]
		if { [llength $foundLines] == 0 } {
			return
		}
		# puts -nonewline "found "; puts $foundLines
		set oldlines [.pw.left.found get 0 end]
		set oldlines2 {}
		foreach line $oldlines {
			lappend oldlines2 [ISO8061ToFilename $line]
		}
		# puts -nonewline "old2 "; puts $oldlines2
		set newlines [intersection $oldlines2 $foundLines]
		# puts -nonewline "new "; puts $newlines
		.pw.left.found delete 0 end
		foreach line $newlines {
			.pw.left.found insert end [filenameToISO8061 $line]
		}
		cancelDialogCommand
	}
	if { [.pw.left.found size] == 0 } {
		return
	}
	showFindDialog "Find Fewer Notes" "Narrow" do
}

proc excludeCommand {} {
	proc do {} {
		global searchString
		if { $searchString eq "" } {
			return
		}
		set foundLines [makeFoundList]
		if { [llength $foundLines] == 0 } {
			return
		}
		set oldlines [.pw.left.found get 0 end]
		set newlines [exclusion $oldlines $foundLines]
		.pw.left.found delete 0 end
		foreach line $newlines {
			.pw.left.found insert end [ISO8061ToFilename $line]
		}
		cancelDialogCommand
	}
	if { [.pw.left.found size] == 0 } {
		return
	}
	showFindDialog "Exclude Notes" "Exclude" do
}

proc makeFoundList {} {
	global searchString
	if { $searchString eq "" } {
		return {}
	}
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
		return {}
	}
	set foundLines [lsort [split $output "\n"]]
	return $foundLines
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
	global activeColor todayColor

	# clear all old decorations
	set defaultcolor [.pw.left.calframe.monthYear cget -activebackground]
	decorate "" $defaultcolor

	# decorate today, if it's being shown
	if { $displayedMonth == $todayMonth && $displayedYear == $todayYear } {
		decorate $todayDay $todayColor
	}

	# decorate currently selected day
	foreach button $[winfo children .pw.left.calframe] {
		decorate $displayedDay $activeColor
	}
}

proc createRowOfDayButtons {row line} {
	global activeColor
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
		button .pw.left.calframe.b$day -text $day -command "setDay $day" -activebackground $activeColor	;# mouse over
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
	saveDisplayedNote
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
		saveDisplayedNote
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
frame .pw.left -padx 8 -pady 8

# create a 7-column frame to hold the header, day names and day buttons
frame .pw.left.calframe
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

listbox .pw.left.found -selectmode single
scrollbar .pw.left.sb -orient vertical -command {.pw.left.found yview}

grid .pw.left.found -row 1 -column 0 -sticky news
grid .pw.left.sb -row 1 -column 1 -sticky ns

.pw.left.found configure -yscrollcommand {.pw.left.sb set}

bind .pw.left.found <<ListboxSelect>> {
	# puts -nonewline "listbox select "
	# puts -nonewline [%W curselection]
	# puts [.pw.left.found get [%W curselection]]
	if { [llength [%W curselection]] == 1 } {
		set i [lindex [%W curselection] 0]
		if { [string is integer $i] } {
			set txt [.pw.left.found get $i]
			set txt [ISO8061ToDate $txt]
			lassign $txt y m d
			setYearMonthDay $y $m $d
		}
	}
}

# right hand side of panedwindow is just a text widget
text .pw.note -wrap word -undo 1 -font $defaultFont

.pw add .pw.left .pw.note
pack .pw -fill both -expand yes

loadDisplayedNote

# https://www.pythontutorial.net/tkinter/tkinter-event-binding/
menu .m -tearoff 0
.m add command -label "Cut" -command {puts "Cut"}
.m add command -label "Copy" -command {puts "Copy"}
.m add command -label "Paste" -command {puts "Paste"}
.m add command -label Save -underline 0 -command saveDisplayedNote -accelerator <Control-s>

bind .pw.note <Button-3> {tk_popup .m %X %Y}
bind . <Control-s> {.m invoke Save}

bind . <F2> {setYearMonthDay $todayYear $todayMonth $todayDay}
bind . <F5> {findCommand}
bind . <F6> {widenCommand}
bind . <F7> {narrowCommand}
bind . <F8> {excludeCommand}

wm protocol . WM_DELETE_WINDOW {
	set response [tk_messageBox -type yesno -message "Really quit?"]
	if {$response eq "yes"} {
		# save displayed note
		destroy .
	}
}