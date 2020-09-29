package cmd

const Help = `
gokeybr is a touch-typing training program

Examples: 

   How to run to train your bash commands typing speed, customized for your commands:

       history | go run main.go -f -

   Or train to type random quote, like on typeracer:
   
       fortune | go run main.go -l 1000 -f -

Key bindings:

   ESC   quit
   C-F   skip forward to the next phrase

Files:
	gokeybr stores log of your training sessions in file ~/.gokeybr_stats_log.jsonl.
	Each line in that file contains timestamp, text, and timeline of one session.
	Timeline is list of values of seconds each character in text was typed.
	Last value in timeline will give session duration.

	Purpose of this file is to be able to compute more detailed stats later.

	
	~/.gokeybr_stats.json is used to store general statistics used to generate training sessions.
`
