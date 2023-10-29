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
set optionIgnoreCase 1
set optionSearchWords 0

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
	if {![string is integer -strict $m]} { error "'$m' is not an integer" }
	global monthNames
	set i [lsearch $monthNames $m]
	return $i
}

proc monthName {m} {
	global monthNames
	set i [lindex $monthNames $m]
	if { $i == -1 } { error "'$m' not found in monthNames" }
	return $i
}

proc filename {y m d} {
	if {![string is integer -strict $y]} { error "'$y' is not an integer" }
	if {![string is integer -strict $m]} { error "'$m' is not an integer" }
	if {![string is integer -strict $d]} { error "'$d' is not an integer" }
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
	return [.pw.right.note edit modified]
}

proc isNoteOld? {} {
	global displayedYear displayedMonth displayedDay
	global todayYear todayMonth todayDay
	if { $displayedYear < $todayYear } {
		return 1
	}
	if { $displayedMonth < $todayMonth } {
		return 1
	}
	if { $displayedDay < $todayDay } {
		return 1
	}
	return 0
}

proc unlockNote {} {
	.pw.right.note configure -state normal
}

proc getNoteText {} {
	# end goes beyond the end of the text, adding a new line, so subtract one character
	# get uses line.char for indexes, eg 1.0
	# lines are numbered from 1
	# chars are numbered from 0
	# see p450
	set txt [.pw.right.note get 1.0 "end - 1 c"]
}

proc setNoteText {txt} {
	.pw.right.note configure -state normal
	.pw.right.note configure -undo false
	.pw.right.note delete 1.0 end
	.pw.right.note insert 1.0 $txt
	.pw.right.note configure -undo true
	.pw.right.note edit modified false
	if [isNoteOld?] {
		.pw.right.note configure -state disabled
	}
	focus .pw.right.note
}

# load/save note (always the displayed note)

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
	if ![isNoteDirty?] {
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
		puts -nonewline $f $txt
		close $f
		puts "saved $fname"
	}
}

# search

proc cancelDialogCommand {} {
	destroy .dialog
}

proc showSearchDialog {title label command} {
	global searchString
	global optionIgnoreCase optionSearchWords
	set searchString ""

	toplevel .dialog

	# hide .dialog while we build it
	wm withdraw .dialog

	ttk::entry .dialog.findText -textvariable searchString
	ttk::checkbutton .dialog.ignoreCase -variable optionIgnoreCase -text "Ignore case"
	ttk::checkbutton .dialog.searchWords -variable optionSearchWords -text "Search whole words"
	ttk::button .dialog.findButton -text $label -command $command
	ttk::button .dialog.cancelButton -text "Cancel" -command {cancelDialogCommand}

	grid config .dialog.findText -column 0 -row 0
	grid config .dialog.findButton -column 1 -row 0
	grid config .dialog.cancelButton -column 2 -row 0
	grid config .dialog.ignoreCase -column 0 -row 2 -sticky w
	grid config .dialog.searchWords -column 0 -row 3 -sticky w

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
		.pw.left.foundframe.found delete 0 end
		foreach line $foundLines {
			.pw.left.foundframe.found insert end [filenameToISO8061 $line]
		}
		cancelDialogCommand
	}
	showSearchDialog "Find Notes" "Find" do
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
		set oldlines {}
		foreach line [.pw.left.foundframe.found get 0 end] {
			lappend oldlines [ISO8061ToFilename $line]
		}
		set newlines [union $oldlines $foundLines]
		.pw.left.foundframe.found delete 0 end
		foreach line $newlines {
			.pw.left.foundframe.found insert end [filenameToISO8061 $line]
		}
		cancelDialogCommand
	}
	if { [.pw.left.foundframe.found size] == 0 } {
		return
	}
	showSearchDialog "Find More Notes" "Widen" do
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
		set oldlines {}
		foreach line [.pw.left.foundframe.found get 0 end] {
			lappend oldlines [ISO8061ToFilename $line]
		}
		# puts -nonewline "old "; puts $oldlines
		set newlines [intersection $oldlines $foundLines]
		# puts -nonewline "new "; puts $newlines
		.pw.left.foundframe.found delete 0 end
		foreach line $newlines {
			.pw.left.foundframe.found insert end [filenameToISO8061 $line]
		}
		cancelDialogCommand
	}
	if { [.pw.left.foundframe.found size] == 0 } {
		return
	}
	showSearchDialog "Find Fewer Notes" "Narrow" do
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
		set oldlines {}
		foreach line [.pw.left.foundframe.found get 0 end] {
			lappend oldlines [ISO8061ToFilename $line]
		}
		set newlines [exclusion $oldlines $foundLines]
		.pw.left.foundframe.found delete 0 end
		foreach line $newlines {
			.pw.left.foundframe.found insert end [ISO8061ToFilename $line]
		}
		cancelDialogCommand
	}
	if { [.pw.left.foundframe.found size] == 0 } {
		return
	}
	showSearchDialog "Exclude Notes" "Exclude" do
}

proc grepOptions {} {
	global optionIgnoreCase optionSearchWords
	# --fixed-strings do not use regular expressions
	# -I exclude binary files
	set lst "--fixed-strings --recursive --files-with-matches --exclude-dir='.*' -I"
	if $optionIgnoreCase {
		lappend lst "--ignore-case"
	} else {
		lappend lst "--no-ignore-case"
	}
	if $optionSearchWords {
		lappend lst "--word-regexp"
	}
	return $lst
}

proc makeFoundList {} {
	global searchString
	if { $searchString eq "" } {
		return {}
	}
	set output ""
	set grepstatus 0
	try {
		# [grepOptions] returns a list
		# use Tcl’s argument expansion syntax to provide the list elements as separate arguments
		# (versions of Tcl prior to 8.5 used the eval command to similar effect)
		# see Tcl and the Tk Toolkit 2nd edition p202-203
		# set output [exec echo grep {*}[grepOptions] $searchString [directory]]
		# puts $output
		set output [exec grep {*}[grepOptions] $searchString [directory]]
		puts -nonewline [llength $output]
		puts " grep [grepOptions] $searchString [directory]"
	} trap CHILDSTATUS {results options} {
    	set grepstatus [lindex [dict get $options -errorcode] 2]
		# puts [string cat "grepstatus " $grepstatus " results " $results]
		# status 0 line selected
		# status 1 no line selected result "child process exited abnormally"
		# status 2 error
	}
	if { $grepstatus != 0 } {
		puts "0 grep [grepOptions] $searchString [directory]"
		return {}
	}
	set foundLines [lsort [split $output "\n"]]
	return $foundLines
}

# calendar

proc numDaysInMonth {month year} {
	if {[expr $month < 1] || [expr $month > 12]} {
		error "Invalid month: $month"
	}

	if { [lsearch "4 6 9 11" $month] != -1 } {
		set n 30
	} elseif { $month == 2 } {
		if { [expr $year % 4 == 0] && [expr $year % 100 != 0] || [expr $year % 400 == 0] } {
			set n 29
		} else {
			set n 28
		}
	} else {
		set n 31
	}

 	return $n
}
# for {set i 1} {$i <=12} {incr i} {
# 	puts [numDaysInMonth $i 2023]
# }

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
	global displayedDay displayedMonth displayedYear
	# puts [string cat "refreshCalendar " $displayedYear " " $displayedMonth " " $displayedDay]

	clearDayButtons
	set calendarLines [split [exec ncal -bh -m $displayedMonth -y $displayedYear -1] "\n"]
	# day names do not change
	# set dayNames [regexp -all -inline {\S+} [lindex $calendarLines 1]]
	.pw.left.calframe.monthYear configure -text [string cat [monthName $displayedMonth] " " $displayedYear]
	createDayButtons $calendarLines
}

proc prevDay {} {
	global displayedDay displayedMonth displayedYear
	set y $displayedYear
	set m $displayedMonth
	set d $displayedDay
	if { $d == 1 } {
		if { $m == 1 } {
			set m 12
			incr y -1
		} else {
			incr m -1
		}
		set d [numDaysInMonth $m $y]
	} else {
		incr d -1
	}
	setYearMonthDay $y $m $d
}

proc nextDay {} {
	global displayedDay displayedMonth displayedYear
	set y $displayedYear
	set m $displayedMonth
	set d $displayedDay
	if { $d == [numDaysInMonth $m $y] } {
		if { $m == 12 } {
			set m 1
			incr y
		} else {
			incr m
		}
		set d 1
	} else {
		incr d
	}
	setYearMonthDay $y $m $d
}

proc prevMonth {} {
	global displayedDay displayedMonth displayedYear
	if { $displayedMonth == 1 } {
		set displayedMonth 12
		incr displayedYear -1
	} else {
		incr displayedMonth -1
	}
	setDay [numDaysInMonth $displayedMonth $displayedYear]
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
	saveDisplayedNote
	set displayedDay $day
	decorateDayButtons
	loadDisplayedNote
}

proc setYearMonthDay {y m d} {
	global displayedDay displayedMonth displayedYear
	if { $y == $displayedYear && $m == $displayedMonth } {
		setDay $d
	} else {
		saveDisplayedNote
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

# user interface

panedwindow .pw -orient horizontal ;#-showhandle 1

# outer frame for all widgets on left hand side of panedwindow
frame .pw.left -padx 8 -pady 16

# calendar, at top of left frame

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

# button bar

frame .pw.left.barframe
grid .pw.left.barframe -pady 8

button .pw.left.barframe.find -text "Find" -command "findCommand"
grid .pw.left.barframe.find -row 0 -column 0
button .pw.left.barframe.widen -text "Or" -command "widenCommand"
grid .pw.left.barframe.widen -row 0 -column 1
button .pw.left.barframe.narrow -text "And" -command "narrowCommand"
grid .pw.left.barframe.narrow -row 0 -column 2
button .pw.left.barframe.exclude -text "Not" -command "excludeCommand"
grid .pw.left.barframe.exclude -row 0 -column 3
button .pw.left.barframe.today -text $todayDay -command {setYearMonthDay $todayYear $todayMonth $todayDay}
grid .pw.left.barframe.today -row 0 -column 4
button .pw.left.barframe.unlock -text "Un" -command "unlockNote"
grid .pw.left.barframe.unlock -row 0 -column 5

# found listbox and scrollbar

frame .pw.left.foundframe
grid .pw.left.foundframe

listbox .pw.left.foundframe.found -selectmode single
grid .pw.left.foundframe.found -row 1 -column 0 -pady 8
scrollbar .pw.left.foundframe.sb -orient vertical -command {.pw.left.foundframe.found yview}
grid .pw.left.foundframe.sb -row 1 -column 1 -sticky ns -pady 8

.pw.left.foundframe.found configure -yscrollcommand {.pw.left.foundframe.sb set}
bind .pw.left.foundframe.found <<ListboxSelect>> {
	# puts -nonewline "listbox select "
	# puts -nonewline [%W curselection]
	# puts [.pw.left.foundframe.found get [%W curselection]]
	if { [llength [%W curselection]] == 1 } {
		set i [lindex [%W curselection] 0]
		if { [string is integer $i] } {
			set txt [.pw.left.foundframe.found get $i]
			set txt [ISO8061ToDate $txt]
			lassign $txt y m d
			setYearMonthDay $y $m $d
		}
	}
}

# button bar

# frame .pw.left.barframe
# grid .pw.left.barframe

# button .pw.left.barframe.find -text "Find" -command "findCommand"
# button .pw.left.barframe.widen -text "Or" -command "widenCommand"
# button .pw.left.barframe.narrow -text "And" -command "narrowCommand"
# button .pw.left.barframe.exclude -text "Not" -command "excludeCommand"
# button .pw.left.barframe.today -text $todayDay -command {setYearMonthDay $todayYear $todayMonth $todayDay}
# button .pw.left.barframe.unlock -text "Un" -command "unlockNote"
# grid .pw.left.barframe.find .pw.left.barframe.widen .pw.left.barframe.narrow .pw.left.barframe.exclude .pw.left.barframe.today .pw.left.barframe.unlock

# right

frame .pw.right
pack .pw.right

text .pw.right.note -wrap word -undo 1 -font $defaultFont
pack .pw.right.note

.pw add .pw.left .pw.right
pack .pw -fill both -expand yes

loadDisplayedNote

# https://www.pythontutorial.net/tkinter/tkinter-event-binding/
menu .m -tearoff 0
# Cut the selected text to the clipboard
# set selected_text [.t get sel.first sel.last]
# .t delete sel.first sel.last
# clipboard clear
# clipboard append $selected_text
# .m add command -label "Cut" -command {puts "Cut"}
# Copy the selected text to the clipboard
# set selected_text [.t get sel.first sel.last]
# clipboard clear
# clipboard append $selected_text
# .m add command -label "Copy" -command {puts "Copy"}
# Get the text from the clipboard
# set clipboard_text [clipboard get]
# Insert the text into the text widget
# .t insert end $clipboard_text
# .m add command -label "Paste" -command {puts "Paste"}
#
# .m add command -label "Cut" -command {tk_textCut %W}
# event add <<Cut>> <Control-x>
# .m add command -label "Copy" -command {tk_textCopy %W}
# event add <<Copy>> <Control-c>
# .m add command -label "Paste" -command {tk_textPaste %W}
# event add <<Paste>> <Control-v>
.m add command -label "Select All" -underline 7 -command {.pw.right.note tag add sel 1.0 end} -accelerator <Control-a>
.m add command -label Save -underline 0 -command saveDisplayedNote -accelerator <Control-s>

bind .pw.right.note <Button-3> {tk_popup .m %X %Y}
# bind . <Control-s> {.m invoke Save}

# TODO Control-left yesterday
# TODO Control-right tomorrow
# TODO investigate why keys here fall though to text widget

# Shortcut keys fall through to an underlying text widget in Tcl/Tk because text widgets are the most common type of widget that accepts keyboard input.
# When a user presses a shortcut key, Tk first checks to see if the current focus widget is a text widget.
# If it is, Tk will send the shortcut key event to the text widget. If the text widget does not handle the shortcut key event, Tk will send the event to the next widget in the focus chain.
#
# This behavior is by design. It allows users to use shortcut keys to perform common operations, such as copying and pasting text, even if the current focus widget is not a text widget.
#
# If you do not want shortcut keys to fall through to an underlying text widget, you can prevent this by overriding the handleEvent() method of the widget that you want to handle the shortcut keys.
# The handleEvent() method is responsible for handling all events that are sent to the widget.
#
# To override the handleEvent() method, you will need to create a new class that inherits from the widget class that you want to handle the shortcut keys.
# In the new class, you will need to override the handleEvent() method and implement your own logic for handling shortcut key events.
#
# Once you have created the new class, you will need to create an instance of it and use it to replace the widget that you do not want to handle shortcut keys.
#
# Here is an example of how to override the handleEvent() method to prevent shortcut keys from falling through to an underlying text widget:
#
# Tcl
# class MyWidget extends TkButton {
#   handleEvent {event} {
    # if {$event == "KeyPress"} {
    #   Handle the shortcut key event.
    # } else {
    #   Call the superclass handleEvent() method to handle the event.
    #   super handleEvent $event
    # }
#   }
# }
bind . <F2> {setYearMonthDay $todayYear $todayMonth $todayDay}
bind . <F4> {unlockNote}

bind . <F5> {findCommand}
bind . <F6> {widenCommand}
bind . <F7> {narrowCommand}
bind . <F8> {excludeCommand}

bind . <F9> {prevDay}
bind . <F10> {nextDay}

wm protocol . WM_DELETE_WINDOW {
	saveDisplayedNote
	destroy .
}

update
set x [expr {([winfo screenwidth .]-[winfo width .])/2}]
set y [expr {([winfo screenheight .]-[winfo height .])/2}]
wm geometry . +$x+$y
