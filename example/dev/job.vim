call ch_logfile('/tmp/vimchannellog', 'w')
let s:target = expand('<sfile>:r') . '.go'

let s:cmd = 'go run ' . s:target
let s:option = {
\   'in_mode': 'json',
\   'out_mode': 'json',
\ }
let g:job = job_start(s:cmd, s:option)
let g:ch = job_getchannel(g:job)

echo job_info(g:job)
echo ch_info(g:ch)

echo 'ch_sendraw: ' . ch_sendraw(g:job, "start!\n")

" echo 'ch_sendraw: ' . ch_sendraw(g:job, "raw msg\n")
" echo 'ch_readraw: ' . ch_readraw(g:job)
