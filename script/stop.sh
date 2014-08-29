#!/bin/bash
: '
    Copyright 2014 JARST, LLC
    
    This file is part of EMP.

    EMP is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
'


# Kill existing process
if [ -f ~/.config/emp/pid ];
then
  echo "Killing emp server (pid="`cat ~/.config/emp/pid`")"
  kill -2 `cat ~/.config/emp/pid`
  rm ~/.config/emp/pid
else
  echo "Server not running, execute start.sh"
fi

