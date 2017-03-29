cobalt
======

Package cobalt provides a toolkit for building web applications and APIs.

It is primarily intended to be used for api web services. It allows the use
of different encoders such as JSON, MsgPack, XML, etc. by implementing the
Coder interface.

Context contains the http request and response writer. It is passed to all
middleware and route handlers. Context contains helper methods for
extracting route parameters from the request URL, methods for decoding the
body of requests, methods for serving encoded responses, and support for
serving templated HTML.

Template support uses some reasonable defaults. These can be changed by
accessing the Templates field of the Cobalt value. To see an example of
cobalt in action check out http://github.com/ardanlabs/cobaltexample
