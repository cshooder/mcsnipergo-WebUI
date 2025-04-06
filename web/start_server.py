#!/usr/bin/env python3

import http.server
import socketserver
import os
import sys
import json
import subprocess
import time
import socket
import threading
import queue
import re

PORT = 8081
DIRECTORY = "dist"  # Serve from the dist directory

# Global process storage
active_processes = {}

def check_port(port):
    """Check if a port is in use"""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        return s.connect_ex(('localhost', port)) == 0

# File paths for configurations
def get_config_path(filename):
    return os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), filename)

class APIHandler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=DIRECTORY, **kwargs)

    def log_message(self, format, *args):
        print(f"[{self.log_date_time_string()}] {format % args}")
    
    def read_config_file(self, filename):
        path = get_config_path(filename)
        try:
            if not os.path.exists(path):
                print(f"Config file '{filename}' not found. Creating empty file.")
                with open(path, 'w') as f:
                    f.write("")
                return ""
            with open(path, 'r') as f:
                return f.read()
        except Exception as e:
            print(f"Error reading config file '{filename}': {e}")
            return ""
    
    def write_config_file(self, filename, data):
        path = get_config_path(filename)
        try:
            with open(path, 'w') as f:
                f.write(data)
            return True
        except Exception as e:
            print(f"Error writing to config file '{filename}': {e}")
            return False
    
    def do_GET(self):
        if self.path == '/api/config/load':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            
            config = {
                'gcAccounts': self.read_config_file('gc.txt'),
                'gpAccounts': self.read_config_file('gp.txt'),
                'msAccounts': self.read_config_file('ms.txt')
            }
            
            self.wfile.write(json.dumps(config).encode())
            return
        
        elif self.path.startswith('/api/console/'):
            # Extract process ID from URL path
            pid = self.path.split('/')[-1]
            
            try:
                pid = int(pid)
                
                if pid in active_processes:
                    process_data = active_processes[pid]
                    
                    # Get the latest output
                    output_lines = []
                    while not process_data['output_queue'].empty():
                        try:
                            line = process_data['output_queue'].get_nowait()
                            output_lines.append(line)
                        except queue.Empty:
                            break
                    
                    # Check if process is still running
                    is_running = process_data['process'].poll() is None
                    
                    self.send_response(200)
                    self.send_header('Content-type', 'application/json')
                    self.end_headers()
                    
                    response = {
                        'output': output_lines,
                        'isRunning': is_running,
                        'username': process_data['username']
                    }
                    
                    if not is_running:
                        exit_code = process_data['process'].poll()
                        response['exitCode'] = exit_code
                    
                    self.wfile.write(json.dumps(response).encode())
                    return
                else:
                    self.send_response(404)
                    self.send_header('Content-type', 'application/json')
                    self.end_headers()
                    self.wfile.write(json.dumps({'error': f'Process with ID {pid} not found'}).encode())
                    return
            except ValueError:
                self.send_response(400)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({'error': 'Invalid process ID'}).encode())
                return
        
        return super().do_GET()
    
    def do_POST(self):
        if self.path == '/api/config/save':
            content_length = int(self.headers.get('Content-Length', 0))
            post_data = self.rfile.read(content_length)
            
            try:
                config = json.loads(post_data.decode('utf-8'))
                
                # Save each config file
                gc_success = self.write_config_file('gc.txt', config.get('gcAccounts', ''))
                gp_success = self.write_config_file('gp.txt', config.get('gpAccounts', ''))
                ms_success = self.write_config_file('ms.txt', config.get('msAccounts', ''))
                
                if gc_success and gp_success and ms_success:
                    response = {'message': 'Account configurations saved successfully!'}
                    self.send_response(200)
                else:
                    response = {'error': 'Failed to save one or more configuration files'}
                    self.send_response(500)
                
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps(response).encode())
                return
            except json.JSONDecodeError:
                self.send_response(400)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({'error': 'Invalid JSON'}).encode())
                return
            except Exception as e:
                self.send_response(500)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({'error': str(e)}).encode())
                return
                
        elif self.path == '/api/snipe':
            content_length = int(self.headers.get('Content-Length', 0))
            post_data = self.rfile.read(content_length)
            
            try:
                data = json.loads(post_data.decode('utf-8'))
                username = data.get('username', '')
                
                if not username:
                    self.send_response(400)
                    self.send_header('Content-type', 'application/json')
                    self.end_headers()
                    self.wfile.write(json.dumps({'error': 'Username cannot be empty'}).encode())
                    return
                
                print(f"Received snipe request for username: {username}")
                
                # Path to the MCsniperGO executable
                mcsnipergo_path = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "mcsnipergo")
                if os.name == 'nt':  # Windows
                    mcsnipergo_path += ".exe"
                
                # Check if executable exists
                if not os.path.exists(mcsnipergo_path):
                    error_msg = (
                        f"MCsniperGO executable not found at {mcsnipergo_path}. "
                        "You need to build the MCsniperGO executable first. "
                        "Please install Go (golang.org) and run 'go build -o mcsnipergo' in the simplesniper directory."
                    )
                    print(error_msg)
                    self.send_response(500)
                    self.send_header('Content-type', 'application/json')
                    self.end_headers()
                    self.wfile.write(json.dumps({'error': error_msg}).encode())
                    return
                
                # Execute MCsniperGO
                try:
                    # Create a queue for storing the output
                    output_queue = queue.Queue()
                    
                    # Start the process
                    process = subprocess.Popen([mcsnipergo_path, "-u", username], 
                                              cwd=os.path.dirname(mcsnipergo_path),
                                              stdout=subprocess.PIPE, 
                                              stderr=subprocess.STDOUT,
                                              text=True,
                                              bufsize=1,
                                              universal_newlines=True)
                    
                    # Store the process information
                    active_processes[process.pid] = {
                        'process': process,
                        'output_queue': output_queue,
                        'username': username
                    }
                    
                    # Start a thread to read the output
                    def read_output():
                        for line in iter(process.stdout.readline, ''):
                            line = line.rstrip()
                            # Print to server console
                            print(f"[MCsniperGO {process.pid}] {line}")
                            # Add to queue
                            output_queue.put(line)
                        
                        # Process has finished
                        print(f"MCsniperGO process {process.pid} finished with exit code {process.poll()}")
                        
                        # Wait to ensure we collect all output
                        process.wait()
                        
                        # Close the output streams
                        if process.stdout:
                            process.stdout.close()
                    
                    # Start the output reader thread
                    threading.Thread(target=read_output, daemon=True).start()
                    
                    self.send_response(200)
                    self.send_header('Content-type', 'application/json')
                    self.end_headers()
                    response = {
                        'message': f"Snipe request for '{username}' is now being processed.",
                        'pid': process.pid
                    }
                    self.wfile.write(json.dumps(response).encode())
                    return
                except Exception as e:
                    self.send_response(500)
                    self.send_header('Content-type', 'application/json')
                    self.end_headers()
                    self.wfile.write(json.dumps({'error': f'Error executing MCsniperGO: {str(e)}'}).encode())
                    return
                
            except json.JSONDecodeError:
                self.send_response(400)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({'error': 'Invalid JSON'}).encode())
                return
            except Exception as e:
                self.send_response(500)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({'error': f'Server error: {str(e)}'}).encode())
                return
            
        self.send_response(404)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps({'error': 'Endpoint not found'}).encode())

def main():
    global PORT
    
    # Check if dist directory exists
    if not os.path.exists(DIRECTORY):
        print(f"Error: '{DIRECTORY}' directory not found!")
        print("Make sure you're running this script from the web directory.")
        sys.exit(1)
    
    # Find an available port if the default is in use
    while check_port(PORT):
        print(f"Port {PORT} is already in use. Trying port {PORT+1}...")
        PORT += 1
        
    print(f"Starting HTTP server on http://localhost:{PORT}")
    print(f"Serving files from '{DIRECTORY}' directory")
    print("Press Ctrl+C to quit")
    
    with socketserver.TCPServer(("", PORT), APIHandler) as httpd:
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\nShutting down server")
            httpd.shutdown()

if __name__ == "__main__":
    main() 