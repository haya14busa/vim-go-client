call ch_logfile('/tmp/vimchannellog', 'w')
let s:target = expand('<sfile>:r') . '.go'

let s:cmd = 'go run ' . s:target
let s:option = {
\   'in_mode': 'json',
\   'out_mode': 'json',
\ }

if !exists('g:job')
  let g:job = job_start(s:cmd, s:option)
endif

if !exists('g:ch')
  let g:ch = job_getchannel(g:job)
endif

echo job_info(g:job)
echo ch_info(g:ch)
echo ch_evalexpr(g:ch, 'hi')

" echo 'ch_sendraw: ' . ch_sendraw(g:job, "start!\n")
" echo 'ch_sendraw: ' . ch_sendraw(g:job, "start!\n")

" echo 'ch_sendraw: ' . ch_sendraw(g:job, "raw msg\n")
" echo 'ch_readraw: ' . ch_readraw(g:job)
