
import sys
import re

def fix_parser(file_path):
    with open(file_path, 'r') as f:
        lines = f.readlines()

    new_lines = []
    
    # Buffers to hold the current function's lines and state
    func_buffer = []
    in_function = False
    has_real_goto = False
    brace_depth = 0
    
    label_re = re.compile(r'^\s*errorExit:\s*$')

    for line in lines:
        stripped = line.strip()
        
        # Simplified function detection
        if line.startswith("func (p *OBParser)"):
            # Flush previous buffer if needed (shouldn't happen if balanced)
            if func_buffer:
                new_lines.extend(func_buffer)
            
            in_function = True
            func_buffer = [line]
            has_real_goto = False
            brace_depth = 1 # We assume the line ends with {
            
            # Double check if line ends with {
            if not stripped.endswith("{"):
                # Should not happen for this file based on observation, but safe to assume 1 if it does
                pass
                
        elif in_function:
            func_buffer.append(line)
            
            # Track braces
            brace_depth += line.count('{')
            brace_depth -= line.count('}')
            
            # Check for real goto
            if stripped.startswith("goto errorExit") and "Trick" not in line:
                has_real_goto = True
            
            if brace_depth == 0:
                # End of function, process the buffer
                processed_buffer = []
                for l in func_buffer:
                    l_stripped = l.strip()
                    # Identify the trick line
                    is_trick = l_stripped.startswith("goto errorExit") and "Trick" in l
                    
                    if is_trick:
                        continue # Always remove the trick line
                    
                    if label_re.match(l) and not has_real_goto:
                        continue # Remove label if no real goto jumps to it
                    
                    processed_buffer.append(l)
                
                new_lines.extend(processed_buffer)
                in_function = False
                func_buffer = []
        else:
            new_lines.append(line)
            
    # Flush remaining
    if func_buffer:
         new_lines.extend(func_buffer)

    with open(file_path, 'w') as f:
        f.writelines(new_lines)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python fix_parser.py <file_path>")
        sys.exit(1)
    
    fix_parser(sys.argv[1])
