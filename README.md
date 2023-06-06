# Commonplace and Incremental Notes - Work in Progress!

There is [a lot of note-taking software out there](https://en.wikipedia.org/wiki/Comparison_of_note-taking_software), and the trouble with most of it is that it's too darn complicated. Which is a shame, because note-taking is something a lot of people do with computers.

There are simple exceptions like [Google Keep](https://keep.google.com/#home) and [thinktype](https://thinktype.app/) but it seems to me most of the big hitters become useless to most people because they require so much investment in learning how to use them.

## Types of note taking

1. [Commonplace books](https://en.wikipedia.org/wiki/Commonplace_book) Leonardo da Vinci kept all of his notes in one big book. If he liked something he put it down. This is known as a commonplace book, and it is about how detailed your note-taking system should be unless you plan on thinking more elaborately than Leonardo da Vinci.
2. [Zettelkasten](https://en.wikipedia.org/wiki/Zettelkasten) or a card file: small items of information stored on paper slips or cards that may be linked to each other through subject headings or other metadata such as numbers and tags.
3. [Incremental notes](https://thesephist.com/posts/inc/) are like a diary, but for notes: start a new page every day and fill it with what you're doing, not doing, or reading, or whatever.

The aim here is to make something that explicitly covers commonplace books and incremental notes, and enable the functionality of zettelkasten using hashtags. After playing around with some designs for a while, I thought it might work to split commonplace books and incremental notes into two similar apps, so that they can be run side-by-side, so text can be cut and pasted between them. The workflow being to start by entering text into the daily note, and then pasting it to a section in a commonplace book if that information endured.

## com

`com` is a simple app that allows you to write and retrieve commonplace book notes. You can have more than one commonplace book; the default one is called `Default`, but you can create others.

The user interface only has four elements:

1. The text of the current note
2. An entry box where you can type for a word that exists in a note (in the current book)
3. A list of note titles that have been found by typing into the entry box above it
4. A toolbar, containing commands to: change the current book or create a new book; create a new note; search for hastags in the book.

## inc

`inc` is a simple app that allows you to write and retrieve incremental notes. You can have more than one notebook; the default one is called `Default`, but you can create others.

The idea came from [The Sephist's article](https://thesephist.com/posts/inc/) and from using [rednotebook](https://rednotebook.app) for a while.

## markdown

## hashtags

## Implementation

`com` and `inc` are written in [Go](https://go.dev/), with the user interface done using the [Fyne](https://fyne.io/) library. The search code is copied and adapted from [Andrew Healey's grup](https://healeycodes.com/beating-grep-with-go). There's no indexing or anything fancy going on under the hood - what we have here is a text editor, grep and a simple user interface.

## Local File Storage

All the notes are stored as text files in a directory tree. The root is `.goldnotebook`.

The commonplace notes are stored in a subtree of that called `com`, which contains a number of directories, one for each book. The default book is called `Default`. Inside each book directory are text files, with the name of the file being a santitized version of the first line of each note. For example, if you had a note on cooking tips in the default book, it would be stored in a file called `.goldnotebook/com/Default/Cooking Tips.txt`.

The sequential notes are stored in a subtree called `inc`, which contains a number of directories, one for each book. The default book is called `Default`. Inside each book directory are directories for each year, and inside each of those, directories for each month. Each month directory contains text files for each day of the month. For example, if you made a note on January 5th 2023 in the default book, it would be stored in a file called `.goldnotebook/inc/Default/2023/01/05.txt`.

You can shadow the entire `.goldnotebook` directory tree in cloud storage, archive them in a [git](https://git-scm.com/) repository (which you can upload to a private github repository), or backup all the notes using, rsync or zip, for example, `zip -r <filename> .goldnotebook`. I use a little bash script to name the backup files after the date they were made, for example:

```bash
today=`date +%Y-%m-%d`
filename="goldnotebook$today.zip"
cd ~
zip -r $filename .goldnotebook
```

## TODO

- Better text editor (including spellchecking, found word highlighting, more visible caret, more keyboard shortcuts)
- The current search is very efficient, but case sensitive
- Support for moving text from `inc` to `com`, to facilitate short term to long term note workflow; maybe right-click popup 'copy selected text to commonplace book' ...
- The 'choose a hashtag' dialog uses a select widget, which should really be a list widget
- Markdown support in `com`
- More support for hashtags?
- Support for creating backups, or git, or cloud (Fyne has some cloud support)
- Many little quality-of-life tweaks

## History

In the early 1990s, I wrote something called [Idealist](https://en.wikipedia.org/wiki/IdeaList). That grew out of the idea of merging database and text editor functionality, and morphed eventually into a package of components that were used to build applications in the museum, archive, and library sectors. I, like many others, just used it to take notes.

Everytime I use a computer, I end up taking notes. Playing a game, developing a new app, doing finances, reading about the worldwide political horrorshows, building bicycles, planning a vegetable garden, following a tv series, reading a book, ... everything seems to generate notes.

So I've spent decades trying different ways of doing that, trying different apps and methodologies, storing files all over the place and eventually losing them, or not being able to transfer them into the shiny new fashionable app. Idealist is too old, only runs on Windows, and I don't have the source, so I can't use that. After years of nagging at myself, I surrendered to the itch and am developing my own new app(s), to fit my needs, and trying really hard to do it in the simplest way possible. Because plain and simple are good.