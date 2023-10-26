parray tcl_platform
# tcl_platform(byteOrder)     = littleEndian
# tcl_platform(engine)        = Tcl
# tcl_platform(machine)       = x86_64
# tcl_platform(os)            = Linux
# tcl_platform(osVersion)     = 6.2.0-35-generic
# tcl_platform(pathSeparator) = :
# tcl_platform(platform)      = unix
# tcl_platform(pointerSize)   = 8
# tcl_platform(threaded)      = 1
# tcl_platform(user)          = gilbert
# tcl_platform(wordSize)      = 8

puts [string cat "tcl_version " $tcl_version]
# tcl_version 8.6

puts [string cat "tcl_patchLevel " $tcl_patchLevel]
# tcl_version 8.6.12

puts [string cat "tcl_library " $tcl_library]
# /usr/share/tcltk/tcl8.6

set paths [split $auto_path " "]
foreach path $paths {
   puts $path
}
# /usr/share/tcltk/tcl8.6
# /usr/share/tcltk
# /usr/lib
# /usr/local/lib/tcltk
# /usr/local/share/tcltk
# /usr/lib/tcltk/x86_64-linux-gnu
# /usr/lib/tcltk
# /usr/lib/tcltk/tcl8.6

puts $[package ifneeded Tk [package require Tk]]
#$load /usr/lib/x86_64-linux-gnu/libtk8.6.so

package require Tk
# 8.6.12

puts [info loaded]
# {/usr/lib/x86_64-linux-gnu/libtk8.6.so Tk}
