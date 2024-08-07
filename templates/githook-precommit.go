package templates

const GitHookPRECOMMIT = `#!/bin/bash

arr=(%s)

for i in "${arr[@]}"
do
    git diff --cached --name-only | xargs grep --with-filename -n $i && echo "COMMIT REJECTED! Found '$i' references. Please remove them before commiting." && exit 1
done

exit 0`
