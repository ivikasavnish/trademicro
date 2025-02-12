#!/usr/bin/env python3
import subprocess
import json
import os
import signal
import argparse
from datetime import datetime
import sys
from typing import Dict, Optional

class ProcessManager:
    def __init__(self, state_file: str = "process_state.json"):
        self.state_file = state_file
        self.processes: Dict[str, dict] = self._load_state()

    def _load_state(self) -> dict:
        """Load process state from JSON file"""
        if os.path.exists(self.state_file):
            with open(self.state_file, 'r') as f:
                return json.load(f)
        return {}

    def _save_state(self):
        """Save current process state to JSON file"""
        with open(self.state_file, 'w') as f:
            json.dump(self.processes, f, indent=4)

    def start_process(self, script_name: str, args: list) -> bool:
        """Start a new trading process"""
        if script_name in self.processes and self.is_running(script_name):
            print(f"Process {script_name} is already running")
            return False

        try:
            cmd = ['python3', script_name] + args
            process = subprocess.Popen(
                cmd,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
                start_new_session=True
            )
            
            self.processes[script_name] = {
                'pid': process.pid,
                'args': args,
                'status': 'running',
                'started_at': datetime.now().isoformat(),
                'command': ' '.join(cmd)
            }
            self._save_state()
            print(f"Started {script_name} with PID {process.pid}")
            return True
        except Exception as e:
            print(f"Error starting process: {e}")
            return False

    def stop_process(self, script_name: str) -> bool:
        """Stop a running process"""
        if script_name not in self.processes:
            print(f"Process {script_name} not found")
            return False

        pid = self.processes[script_name]['pid']
        try:
            os.killpg(os.getpgid(pid), signal.SIGTERM)
            self.processes[script_name]['status'] = 'stopped'
            self._save_state()
            print(f"Stopped process {script_name} (PID: {pid})")
            return True
        except ProcessLookupError:
            print(f"Process {pid} not found - may have already terminated")
            self.processes[script_name]['status'] = 'stopped'
            self._save_state()
            return True
        except Exception as e:
            print(f"Error stopping process: {e}")
            return False

    def resume_process(self, script_name: str) -> bool:
        """Resume a stopped process"""
        if script_name not in self.processes:
            print(f"Process {script_name} not found")
            return False

        if self.is_running(script_name):
            print(f"Process {script_name} is already running")
            return False

        return self.start_process(script_name, self.processes[script_name]['args'])

    def list_processes(self):
        """List all managed processes and their status"""
        print("\nManaged Processes:")
        print("-" * 80)
        print(f"{'Script Name':<20} {'PID':<10} {'Status':<10} {'Started At':<25} {'Command'}")
        print("-" * 80)
        
        for script_name, info in self.processes.items():
            status = 'running' if self.is_running(script_name) else 'stopped'
            print(f"{script_name:<20} {info['pid']:<10} {status:<10} {info['started_at']:<25} {info['command']}")

    def is_running(self, script_name: str) -> bool:
        """Check if a process is currently running"""
        if script_name not in self.processes:
            return False
        
        try:
            pid = self.processes[script_name]['pid']
            os.kill(pid, 0)
            return True
        except ProcessLookupError:
            return False
        except Exception:
            return False

def main():
    parser = argparse.ArgumentParser(description='Trading Process Manager')
    parser.add_argument('action', choices=['start', 'stop', 'resume', 'list'],
                       help='Action to perform')
    parser.add_argument('script', nargs='?', help='Script name (not required for list)')
    parser.add_argument('args', nargs=argparse.REMAINDER, 
                       help='Additional arguments for the script')

    args = parser.parse_args()
    pm = ProcessManager()

    if args.action == 'list':
        pm.list_processes()
    elif args.action in ['start', 'stop', 'resume']:
        if not args.script:
            print("Script name is required for this action")
            sys.exit(1)
        
        action_map = {
            'start': pm.start_process,
            'stop': pm.stop_process,
            'resume': pm.resume_process
        }
        
        success = action_map[args.action](args.script, args.args)
        if not success:
            sys.exit(1)

if __name__ == "__main__":
    main()
