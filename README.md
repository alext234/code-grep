A web-based application for browsing source code which supports definition and reference search (think LXR for linux kernel).

The application used to be hosted on Google Compute Engine at url code-grep.com, but has been shut down. 

Here are some demo videos of the application:

Check out source code from github and start browsing:
https://www.youtube.com/watch?v=x7MWOKa0GkQ

Browse Linux Kernel on code-grep.com:
https://www.youtube.com/watch?v=E-QxI-5AyZo

Front end: angularjs, highlightjs, and some adhoc javascript.

Back end: webserver and backend workers in Golang; database is MongoDB.

Deployment is based on Ansible.  

