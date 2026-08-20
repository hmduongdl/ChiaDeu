import re
import os

for f in ['pkg/serverless/groupsexpensesdetail.go', 'pkg/serverless/groupsexpensesindex.go']:
    with open(f, 'r') as file:
        content = file.read()
    
    # Remove unused imports
    for i in ['"errors"', '"strings"', '"time"', '"github.com/hmduongdl/ChiaDeu/pkg/expenses"', '"github.com/hmduongdl/ChiaDeu/models"', '"github.com/hmduongdl/ChiaDeu/services"']:
        content = content.replace(i, '')
        
    with open(f, 'w') as file:
        file.write(content)
