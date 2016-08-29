call ch_logfile('/tmp/vimchannellog', 'w')
let s:target = expand('<sfile>:r') . '.go'
let s:cmd = 'go run ' . s:target
" let s:target = expand('<sfile>:r')
" let s:cmd = s:target

function! s:err_cb(...) abort
  echom '---err_cb---'
  echom string(a:000)
endfunction

let s:option = {
\   'in_mode': 'json',
\   'out_mode': 'json',
\
\   'err_cb': function('s:err_cb'),
\ }

if !exists('g:job')
  let g:job = job_start(s:cmd, s:option)
endif

if !exists('g:ch')
  let g:ch = job_getchannel(g:job)
endif

echo job_info(g:job)
echo ch_info(g:ch)

function! s:cb(...) abort
  echom '---cb---'
  echom string(a:000)
endfunction

call ch_evalexpr(g:ch, 'hi')
call ch_evalexpr(g:job, 'hi')
call ch_sendexpr(g:ch, 'hi', {'callback': function('s:cb')})
call ch_sendexpr(g:job, 'hi', {'callback': function('s:cb')})
