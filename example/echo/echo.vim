call ch_logfile('/tmp/vimchannellog', 'w')
let s:target = expand('<sfile>:r') . '.go'
let s:cmd = 'go run ' . s:target
let s:option = {
\   'in_mode': 'json',
\   'out_mode': 'json',
\ }
let s:job = job_start(s:cmd, s:option)
echom ch_evalexpr(s:job, 'hi!')
" => hi!

let s:done = 0

function! s:cb(ch, msg) abort
  let s:done = 1
  echom string(a:msg)
endfunction

call ch_sendexpr(s:job, {'msg': 'hi!'}, {'callback': function('s:cb')})

while 1
  if s:done
    call job_stop(s:job)
    break
  endif
  sleep 10ms
endwhile
