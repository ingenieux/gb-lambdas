# gb plugin for lambdas

Uses [eawsy/aws-lambda-go-shim](github.com/eawsy/aws-lambda-go-shim)

## Building

Requirements: make, Docker, golang, and [gb](https://github.com/constabulary/gb)

Steps:

  * clone, checkout, cd
  * ```gb vendor restore```
  * ```make clean all```
  * ```install -m 755 bin/gb-lambdas ~/bin # or somewhere in your path```

## Using

See [gb-lambdas-sample](https://github.com/ingenieyux/gb-lambdas-sample)

## Troubleshooting:

Run ```gb-lambdas``` with ```GB_LAMBDAS_LOGLEVEL``` set to ```debug```:

```
$ GB_LAMBDAS_LOGLEVEL="debug" gb lambdas
```
