This is an example code that reports AIDA64 metics to pushgateway through WMI and works fine on my machine.

Note that it doesn't expose an api for prometheus to scrape directly, but you can implement one if you want to.

How to use:
 - modify the pushgateway url constant.
 - run go build -ldflags="-H windowsgui" ./wmi_exporter.go
 - run the executable or add it as a service.
