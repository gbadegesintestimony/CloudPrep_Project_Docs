package curriculum

import "cloudprep-backend/internal/models"

type ExamPresetInfo struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	ModuleScope      string   `json:"moduleScope"`
	TopicIDs         []string `json:"topicIds"`
	QuestionCount    int      `json:"questionCount"`
	TimeLimitMinutes int      `json:"timeLimitMinutes"`
}

func GetExamPresets() []ExamPresetInfo {
	return []ExamPresetInfo{
		{
			ID:               models.PresetModule1,
			Title:            "Module 1 Mock Exam — Cloud & DevOps Foundations",
			Description:      "Timed assessment covering Introduction to Computing, Linux Fundamentals, Bash Scripting & Automation, and Git & Version Control across definitions, CLI commands, comparisons, and scenarios.",
			ModuleScope:      "Module 1 (Weeks 1–4)",
			TopicIDs:         []string{"topic-1-cloud-intro", "topic-2-linux", "topic-3-bash", "topic-4-git"},
			QuestionCount:    30,
			TimeLimitMinutes: 40,
		},
		{
			ID:               models.PresetModule2,
			Title:            "Module 2 Mock Exam — Cloud Architecture, Services & AI",
			Description:      "Timed assessment covering Core Compute & IAM, Cloud Networking, Cloud Storage, Databases, Serverless & Application Services, and Cloud AI & MLOps.",
			ModuleScope:      "Module 2 (Weeks 1–4)",
			TopicIDs:         []string{"topic-5-compute-iam", "topic-6-networking", "topic-7-storage", "topic-8-databases", "topic-9-serverless", "topic-10-ai-mlops"},
			QuestionCount:    40,
			TimeLimitMinutes: 55,
		},
		{
			ID:               models.PresetComprehensiveFinal,
			Title:            "Comprehensive Final Mock Exam — All 10 Curriculum Topics",
			Description:      "Full-length Cloud Engineering mock examination spanning all 10 curriculum topics from Module 1 and Module 2 with balanced question styles and dynamically shuffled options.",
			ModuleScope:      "Module 1 & Module 2 (All 10 Topics)",
			TopicIDs: []string{
				"topic-1-cloud-intro", "topic-2-linux", "topic-3-bash", "topic-4-git",
				"topic-5-compute-iam", "topic-6-networking", "topic-7-storage", "topic-8-databases",
				"topic-9-serverless", "topic-10-ai-mlops",
			},
			QuestionCount:    60,
			TimeLimitMinutes: 80,
		},
	}
}

// GetCurriculumTopics returns the 10 structured curriculum topics and learning objectives extracted from school notes/.
func GetCurriculumTopics() []models.CurriculumTopic {
	return []models.CurriculumTopic{
		{
			ID:           "topic-1-cloud-intro",
			Slug:         "introduction-to-computing",
			ModuleNumber: 1,
			ModuleTitle:  "Module 1: Cloud & DevOps Foundations",
			WeekCode:     "M1_W1",
			Title:        "1. Introduction to Computing",
			Description:  "Cloud computing definitions, traditional on-premises vs. cloud infrastructure, service models (IaaS, PaaS, SaaS), deployment models (Public, Private, Hybrid, Multi-cloud), the Shared Responsibility Model, SDLC integration, cloud sub-occupations, and the Seven 'ANDS' of Cloud Engineering.",
			SourcePdf:    "M1_W1_Introduction to Cloud Computing.pdf",
			Ordering:     1,
			Objectives: []models.LearningObjective{
				{
					ID:              "obj-1-1",
					TopicID:         "topic-1-cloud-intro",
					Code:            "M1.W1.OBJ1",
					Title:           "Cloud Computing Definition & On-Premises vs. Cloud Comparison",
					Statement:       "Define cloud computing and compare traditional on-premises infrastructure with cloud infrastructure.",
					SourceReference: "Module 1, Week 1 — Introduction to Cloud Computing (Slides 5, 9)",
					Ordering:        1,
					CurriculumExcerpt: `The cloud is a global network of remote servers housed in massive data centers (AWS, Microsoft Azure, Google Cloud Platform, DigitalOcean, Heroku, Huawei, Alibaba). Cloud computing is the on-demand delivery of computing resources — servers, storage, databases, networking, software, analytics, and intelligence — over the internet to offer faster innovation, flexible resources, and economies of scale.
Traditional (On-Premises) Infrastructure: You buy physical servers and store them in your own data center or server room. You pay for everything upfront (hardware, power, cooling, space, staff). You provision capacity based on peak expected load, paying for idle resources most of the time, or waiting weeks/months to buy hardware if you underestimate.
Cloud Infrastructure: You rent computing resources from a provider, pay only for what you use (pay-as-you-go), and scale up or down in minutes. The provider handles physical infrastructure, redundancy, and data center operations.`,
				},
				{
					ID:              "obj-1-2",
					TopicID:         "topic-1-cloud-intro",
					Code:            "M1.W1.OBJ2",
					Title:           "Cloud Service Models (IaaS, PaaS, SaaS) & Shared Responsibility",
					Statement:       "Differentiate between IaaS, PaaS, and SaaS, identify real-world examples, and apply the Shared Responsibility Model.",
					SourceReference: "Module 1, Week 1 — Introduction to Cloud Computing (Slides 2, 6, 14)",
					Ordering:        2,
					CurriculumExcerpt: `Cloud Engineering involves designing, building, and managing cloud-based systems and services.
1. Infrastructure as a Service (IaaS): Provides virtualized computing resources over the internet (e.g., AWS EC2, Google Compute Engine, Azure Virtual Machines, DigitalOcean Droplets). In an IaaS setup, the customer is responsible for patching and managing the guest Operating System, applications, and data, while the provider manages the physical hardware, storage, networking, and virtualization layer (Shared Responsibility Model).
2. Platform as a Service (PaaS): Provides hardware and software tools over the internet for application development and hosting without managing the underlying OS (e.g., Google App Engine, Azure App Service, Heroku, AWS Elastic Beanstalk).
3. Software as a Service (SaaS): Delivers ready-to-use software applications over the internet to end users (e.g., Google Workspace, Microsoft Office 365).`,
				},
				{
					ID:              "obj-1-3",
					TopicID:         "topic-1-cloud-intro",
					Code:            "M1.W1.OBJ3",
					Title:           "Cloud Deployment Models (Public, Private, Hybrid, Multi-Cloud)",
					Statement:       "Explain the four major cloud deployment models and match them to organizational constraints.",
					SourceReference: "Module 1, Week 1 — Introduction to Cloud Computing (Slide 8)",
					Ordering:        3,
					CurriculumExcerpt: `1. Public Cloud: Infrastructure owned and operated by a third-party cloud provider (AWS, Azure, GCP) and shared across multiple tenants over the internet.
2. Private Cloud: Services maintained on a private network (e.g., VMware, OpenStack) with infrastructure dedicated to a single organization (hosted on-premises or by a third party). Provides maximum control and data isolation at higher cost; common in banking and government.
3. Hybrid Cloud: Combines public and private clouds, allowing data and applications to be shared between them.
4. Multi-cloud: Using cloud services from more than one public cloud provider (e.g., AWS for compute + GCP for AI/ML). Avoids vendor lock-in but increases operational complexity.`,
				},
				{
					ID:              "obj-1-4",
					TopicID:         "topic-1-cloud-intro",
					Code:            "M1.W1.OBJ4",
					Title:           "SDLC, Cloud Sub-Occupations & The Seven 'ANDS' of Cloud Engineering",
					Statement:       "Situate cloud engineering within the SDLC and identify key cloud sub-occupations and essential tool categories.",
					SourceReference: "Module 1, Week 1 — Introduction to Cloud Computing (Slides 4, 7, 10–13)",
					Ordering:        4,
					CurriculumExcerpt: `SDLC phases: Planning -> Design -> Development -> Testing -> Deployment -> Maintenance. Cloud engineering transforms the SDLC from "Build -> Deploy -> Hope it works" into "Build -> Test at scale -> Deploy automatically -> Monitor -> Improve continuously."
Sub-Occupations:
- Cloud Architect: Designs cloud infrastructure aligned with business goals (Tools: AWS CloudFormation, Azure Resource Manager, Google Cloud Deployment Manager).
- Cloud DevOps Engineer: Implements CI/CD pipelines and automation (Tools: Jenkins, Bitbucket Pipelines, GitHub Actions, Terraform, Ansible).
- Cloud Security Engineer: Secures cloud environments, access controls, compliance (Tools: AWS Security Hub, Azure Security Center, GCP Security Command Center).
- Cloud Network Engineer: Designs/tunes cloud networks (Tools: AWS VPC, Azure Virtual Network, GCP VPC).
- Cloud Data Engineer: Data warehousing, ETL, analytics (Tools: AWS Redshift, Google BigQuery, Azure Synapse Analytics).
The Seven "ANDS" / Tool Categories: (1) Infrastructure as Code (Terraform, CloudFormation, ARM) & Configuration Management (Ansible, Chef, Puppet), (2) Communication & Collaboration (Slack, Jira, Linear, Notion), (3) CI/CD (Jenkins, GitHub Actions, GitLab CI/CD, ArgoCD, CircleCI), (4) Monitoring and Logging (AWS CloudWatch, Azure Monitor, GCP Operations Suite, Prometheus, Grafana), (5) Containerisation & Orchestration (Docker, Kubernetes, Docker Swarm, AWS EKS, AWS ECS), (6) Networking & Security (AWS VPC, SonarQube, Azure VPN), (7) AI & Machine Learning.`,
				},
			},
		},
		{
			ID:           "topic-2-linux",
			Slug:         "linux-fundamentals",
			ModuleNumber: 1,
			ModuleTitle:  "Module 1: Cloud & DevOps Foundations",
			WeekCode:     "M1_W2",
			Title:        "2. Linux Fundamentals",
			Description:  "Operating systems, Linux kernel vs. shell, file system hierarchy (/bin, /sbin, /etc, /var, /tmp, /usr, /dev), core file navigation commands, users & groups (/etc/passwd, /etc/shadow, /etc/group, usermod -aG), file permissions (chmod, chown, octal 755/640), process management (daemon, orphan, zombie, ps, top, kill), apt package management, SSH, and disk mounting.",
			SourcePdf:    "M1_W2_Linux Fundamentals.pdf",
			Ordering:     2,
			Objectives: []models.LearningObjective{
				{
					ID:              "obj-2-1",
					TopicID:         "topic-2-linux",
					Code:            "M1.W2.OBJ1",
					Title:           "OS Concepts, Linux Kernel vs. Shell & File System Hierarchy",
					Statement:       "Explain Operating System fundamentals, distinguish kernel vs. shell, and navigate the Linux directory tree and core file commands.",
					SourceReference: "Module 1, Week 2 — Linux Fundamentals (Slides 2–6, 17, 20)",
					Ordering:        1,
					CurriculumExcerpt: `An Operating System (OS) is the intermediary bridge between hardware (CPU, memory, disk) and software applications. Linux is an open-source OS created by Linus Torvalds in 1991.
- Kernel: The core part of Linux that talks directly to hardware (like the engine of a car).
- Shell: The program running in the terminal that interprets and processes commands (bash, zsh, ksh, fish) (like the dashboard and steering wheel).
- Linux Distributions (distros): Ubuntu, Debian, Fedora, CentOS, Rocky Linux, AlmaLinux, Red Hat Enterprise Linux (RHEL), Amazon Linux, Kali.
Core Commands: pwd (show current directory), ls (list files; ls -a includes hidden files; ls -l long format with permissions), cd (change directory), mkdir (create directory), touch (create empty file), rm (remove file), mv (move or rename file), cp (copy file), clear, cat (view entire file), less (scroll file), head -n 5 (first 5 lines), tail -n 5 (last 5 lines), man <cmd> (manual pages), <cmd> --help.
Text Editors: nano (Ctrl+O to save, Ctrl+X to exit), vi/vim ('i' to insert, 'Esc' to exit insert mode, ':wq' to save and quit).
Linux File System Hierarchy:
- / (Root directory): top of the unified filesystem tree.
- /bin (Essential user binaries): ls, cat, cp, mv.
- /sbin (System administration binaries, mostly for root): ifconfig, reboot, fdisk.
- /etc (System-wide configuration files): /etc/passwd, /etc/fstab.
- /home (Personal user directories): e.g. /home/alice.
- /var (Variable data that changes frequently): logs (/var/log), spool files, caches.
- /tmp (Temporary files): often cleared on reboot.
- /usr (User programs, libraries, and documentation): /usr/bin, /usr/lib, /usr/share.
- /dev (Device files representing hardware): /dev/sda, /dev/tty.`,
				},
				{
					ID:              "obj-2-2",
					TopicID:         "topic-2-linux",
					Code:            "M1.W2.OBJ2",
					Title:           "Users, Groups & Account Files (/etc/passwd, /etc/shadow, /etc/group)",
					Statement:       "Manage Linux users and primary/secondary groups and interpret /etc/passwd entries.",
					SourceReference: "Module 1, Week 2 — Linux Fundamentals (Slides 7–12)",
					Ordering:        2,
					CurriculumExcerpt: `User info is stored in /etc/passwd (7 colon-separated fields: Username : Password Placeholder 'x' : UID : Primary GID : GECOS comment : Home Directory : Login Shell). Actual encrypted password hashes are stored in /etc/shadow.
- Human users typically have UID >= 1000 and login shell /bin/bash.
- System/service users (e.g., systemd-network, www-data, mysql) have UID < 1000, non-home paths like /run/systemd, and shell /usr/sbin/nologin or /bin/false for security so services cannot open interactive shells.
Groups: Each user has exactly one Primary Group (GID in /etc/passwd, controls default group ownership of new files) and zero or more Secondary/Supplementary Groups (listed in /etc/group).
User & Group Commands:
- whoami: print current effective username.
- id [user]: show UID, GID, and all group memberships with numeric IDs.
- groups [user]: list all groups a user belongs to.
- sudo adduser john: high-level interactive command that automatically creates /home/john, sets shell, and prompts for password.
- sudo useradd john: low-level command that creates user without home directory or password unless flags like sudo useradd -m john and sudo passwd john are used.
- groupadd cloud: create a new group.
- usermod -aG developers alice: safely append (-a) user alice to supplementary group (-G) developers without removing existing groups.
- WARNING: usermod -G developers alice (without -a) overwrites all secondary groups, which can accidentally remove sudo access!
- gpasswd -d alice devs: remove alice from secondary group devs.
- userdel / groupdel: delete a user or group.
- su - alice: switch user to alice and start a fresh login shell.`,
				},
				{
					ID:              "obj-2-3",
					TopicID:         "topic-2-linux",
					Code:            "M1.W2.OBJ3",
					Title:           "File Permissions (chmod) & Ownership (chown)",
					Statement:       "Configure and interpret Linux symbolic and octal permissions and file ownership.",
					SourceReference: "Module 1, Week 2 — Linux Fundamentals (Slides 13–14, 23)",
					Ordering:        3,
					CurriculumExcerpt: `Permissions apply to 3 categories: Owner (u), Group (g), and Others (o) (or All = a).
- r (Read = 4): view file contents; on a directory, list directory contents (ls).
- w (Write = 2): modify file contents; on a directory, create/delete/rename files inside it.
- x (Execute = 1): run a file as a program/script; on a directory, enter (cd) into the directory.
Octal values: 7 = rwx (4+2+1), 6 = rw- (4+2), 5 = r-x (4+1), 4 = r--, 0 = ---.
Example: -rw-r--r-- (644) means Owner has read+write, Group has read-only, Others have read-only. chmod 640 test.txt sets rw-r-----.
Symbolic chmod examples:
- chmod u+x script.sh (add execute for owner)
- chmod g-w report.txt (remove write from group)
- chmod o=r file.txt (set others to read-only)
- chmod a+x run.sh (add execute to user, group, and others)
Changing Ownership (chown):
- chown alice file.txt (change owner only)
- chown bob:devs script.sh (change both owner and group)
- chown :marketing report.pdf (change group only)`,
				},
				{
					ID:              "obj-2-4",
					TopicID:         "topic-2-linux",
					Code:            "M1.W2.OBJ4",
					Title:           "Process Management, Package Managers (APT), SSH & Mounting Storage",
					Statement:       "Manage Linux processes and states, install software with package managers, connect via SSH, and mount filesystems.",
					SourceReference: "Module 1, Week 2 — Linux Fundamentals (Slides 15–16, 18–19, 21–22)",
					Ordering:        4,
					CurriculumExcerpt: `Process Types:
- Parent & Child: processes spawn child processes in a tree (viewed via pstree).
- Daemon: background process detached from terminal waiting for system events/requests.
- Orphaned: parent terminated while child still runs; adopted by systemd/init (PID 1).
- Zombie: process finished execution (dead, uses 0 CPU/RAM) but remains in the process table because its parent has not read its exit status yet. Too many zombies exhaust process table slots.
Process States: New, Ready (in memory waiting for CPU), Running, Sleeping (S), Blocked/Waiting (waiting for I/O), Terminated.
Process Commands: ps (e.g., ps aux), top (real-time monitor), htop (interactive monitor), kill <PID> (sends SIGTERM by default), killall <name>, bg (resume in background), fg (bring to foreground), sleep 100 & (& runs in background).
Package Managers:
- Debian/Ubuntu: apt / dpkg. Commands: sudo apt update (refresh package index), sudo apt install nginx, apt-cache search/show nginx, sudo apt remove nginx (removes binary, keeps configs), sudo apt purge nginx (removes binary AND configuration files).
- RHEL/CentOS/Fedora: yum / dnf.
SSH (Secure Shell): Encrypted remote login protocol. Command: ssh user@ip_address (or ssh -i key.pem user@ip). Verifies host identity and establishes an encrypted tunnel.
Mounting Storage: Linux has a single root tree (no C: or D: drive letters). Workflow to create/mount a virtual disk:
1. fallocate -l 100M disk.img (allocate 100MB file)
2. mkfs.ext4 disk.img (format with ext4 filesystem)
3. mkdir /mnt/fakedisk (create mount point directory)
4. sudo mount disk.img /mnt/fakedisk (attach filesystem)
5. df -h /mnt/fakedisk or mount | grep fakedisk (verify)
6. sudo umount /mnt/fakedisk (safely detach/unmount)`,
				},
			},
		},
		{
			ID:           "topic-3-bash",
			Slug:         "bash-scripting-and-automation",
			ModuleNumber: 1,
			ModuleTitle:  "Module 1: Cloud & DevOps Foundations",
			WeekCode:     "M1_W4",
			Title:        "3. Bash Scripting & Automation",
			Description:  "Bourne Again Shell (Bash) vs. other shells, shebang (#!/bin/bash), variables & positional arguments ($1, $@, $#), read input, conditionals (-eq, -ne, -gt, -lt, -ge, -le, -f, -d, -e), for/while loops, functions, debugging flags (set -x, set -e, set -u, set -euo pipefail), and Cron scheduling (crontab -e/-l/-r, 5-star syntax).",
			SourcePdf:    "M1_ W4_Bash Scripting & Automation.pdf",
			Ordering:     3,
			Objectives: []models.LearningObjective{
				{
					ID:              "obj-3-1",
					TopicID:         "topic-3-bash",
					Code:            "M1.W4.OBJ1",
					Title:           "Bash Fundamentals, Shebang, Variables & Runtime Arguments",
					Statement:       "Write executable Bash scripts using the shebang, variable assignment rules, user input, and positional arguments.",
					SourceReference: "Module 1, Week 4 — Bash Scripting, Automation & Cron (Slides 2–7)",
					Ordering:        1,
					CurriculumExcerpt: `Bash (Bourne Again Shell) is the default shell on most Linux distributions. Other shells include sh (original Unix shell), zsh (default on macOS), fish, ksh, dash (lightweight Debian script shell), ash (embedded), and PowerShell.
Rules for Bash Scripts:
- Must begin with the shebang on line 1: #!/bin/bash (tells the OS kernel to execute the file using /bin/bash).
- Make executable with chmod +x myscript.sh and run via ./myscript.sh or bash myscript.sh.
- Variable assignment MUST NOT have spaces around '=': NAME="Mayowa" is valid; NAME = "Mayowa" fails because Bash treats NAME as a command.
- Access variables with $: echo "Hello, $NAME". Command substitution uses $(date).
- User input: read NAME reads a line from standard input into variable NAME.
- Runtime Positional Arguments:
  - $1, $2, ... : first, second argument passed to the script.
  - $@ : all arguments passed to the script.
  - $# : total number of arguments passed to the script.`,
				},
				{
					ID:              "obj-3-2",
					TopicID:         "topic-3-bash",
					Code:            "M1.W4.OBJ2",
					Title:           "Conditionals, Comparison Operators, Loops & Functions",
					Statement:       "Construct Bash conditionals, numeric/string/file test operators, for/while loops, and reusable functions.",
					SourceReference: "Module 1, Week 4 — Bash Scripting, Automation & Cron (Slides 8–10, 12)",
					Ordering:        2,
					CurriculumExcerpt: `Bash Conditionals (if / then / else / fi):
Numeric comparison operators inside [ ... ]:
- -eq (equal), -ne (not equal), -gt (greater than), -lt (less than), -ge (greater than or equal), -le (less than or equal).
String comparison operators: = (equal) and != (not equal).
File test operators:
- -f "/path" (true if path exists and is a regular file)
- -d "/path" (true if path exists and is a directory)
- -e "/path" (true if path exists, either file or directory)
Loops:
- For loop over list/glob/range: for i in 1 2 3 4 5; do echo "$i"; done or for file in *.txt; do ... done or for i in {1..5}; do mkdir "folder_$i"; done.
- While loop: COUNT=1; while [ $COUNT -le 5 ]; do echo "$COUNT"; COUNT=$((COUNT + 1)); done.
Functions:
backup_file() {
  cp "$1" "$1.bak"
  echo "Backup created: $1.bak"
}
backup_file "config.txt" (inside a function, $1 refers to the function's first argument).`,
				},
				{
					ID:              "obj-3-3",
					TopicID:         "topic-3-bash",
					Code:            "M1.W4.OBJ3",
					Title:           "Debugging Bash Scripts (set -e, set -u, set -x, set -euo pipefail)",
					Statement:       "Apply Bash safety and debugging flags to prevent silent failures and trace script execution.",
					SourceReference: "Module 1, Week 4 — Bash Scripting, Automation & Cron (Slides 6, 11)",
					Ordering:        3,
					CurriculumExcerpt: `Debugging & Safety Flags in Bash:
- set -x (trace mode): prints each command and its expanded arguments to the terminal before executing it.
- set -e (exit on error): stops the script immediately if any command exits with a non-zero status, preventing silent cascading failures.
- set -u (unset variables): treats referencing an unset/undefined variable as an error, catching variable name typos immediately.
- set -euo pipefail: production best practice combining -e, -u, and -o pipefail (which causes a pipeline like cmd1 | cmd2 to fail if any command in the pipe fails, not just the last one).`,
				},
				{
					ID:              "obj-3-4",
					TopicID:         "topic-3-bash",
					Code:            "M1.W4.OBJ4",
					Title:           "Cron & Crontab Task Scheduling (5-Star Syntax)",
					Statement:       "Schedule automated background tasks using crontab commands and 5-field cron expressions.",
					SourceReference: "Module 1, Week 4 — Bash Scripting, Automation & Cron (Slides 13–19)",
					Ordering:        4,
					CurriculumExcerpt: `Cron Terminology:
- Cron: the background system daemon (service) that wakes up every minute to check for scheduled work.
- Crontab (cron table): the file containing the list of scheduled instructions.
- CronJob: a single scheduled task entry (line) inside a crontab.
Crontab Commands:
- crontab -e : edit the current user's crontab in the default editor.
- crontab -l : list/display the current user's active crontab entries.
- crontab -r : remove/delete the user's entire crontab file.
The 5-Field Crontab Syntax (* * * * * command_to_execute):
1. Minute (0–59)
2. Hour (0–23)
3. Day of Month (1–31)
4. Month (1–12 or JAN–DEC)
5. Day of Week (0–6, where 0 is Sunday and 6 is Saturday)
Special Characters:
- * (Asterisk): every possible value (e.g., every hour).
- , (Comma): list of specific values (e.g., 1,15 in Minute = at minute 1 and minute 15).
- - (Hyphen): inclusive range (e.g., 9-17 in Hour = every hour from 9 AM to 5 PM; 1-5 in Day of Week = Monday through Friday).
- / (Slash): step interval (e.g., */10 in Minute = every 10 minutes; 0 */2 * * * = every 2 hours on the hour).
Examples:
- 0 6 * * * : every day at 6:00 AM.
- 2 9 * * * : every day at 9:02 AM.
- 30 14 * * 1-5 : at 2:30 PM on weekdays (Monday–Friday).
Best Practice: Always use absolute paths (e.g., /usr/bin/python3 /home/user/script.py) in cron jobs because cron runs in a minimal environment without your interactive PATH.`,
				},
			},
		},
		{
			ID:           "topic-4-git",
			Slug:         "git-and-version-control",
			ModuleNumber: 1,
			ModuleTitle:  "Module 1: Cloud & DevOps Foundations",
			WeekCode:     "M1_W3",
			Title:        "4. Git & Version Control",
			Description:  "Git vs. GitHub, local configuration (git config --global), authentication (Personal Access Tokens vs. ed25519 SSH keys), working directory vs. staging area vs. repository, core workflow (init, clone, add, commit, status, log, push, pull, fetch, fork), branching, HEAD pointer, and merge conflicts.",
			SourcePdf:    "M1_W3_Git & Version Control - Google Slides.pdf",
			Ordering:     4,
			Objectives: []models.LearningObjective{
				{
					ID:              "obj-4-1",
					TopicID:         "topic-4-git",
					Code:            "M1.W3.OBJ1",
					Title:           "Git vs. GitHub, Identity Configuration & SSH / PAT Authentication",
					Statement:       "Distinguish Git from GitHub, configure global commit identity, and authenticate via SSH keys or Personal Access Tokens.",
					SourceReference: "Module 1, Week 3 — Git & Version Control (Slides 2–6)",
					Ordering:        1,
					CurriculumExcerpt: `Git vs. GitHub:
- Git: a distributed version control system (tool) that runs locally on your machine to track file changes over time. Git works without GitHub.
- GitHub (along with GitLab and Bitbucket): a cloud hosting platform built on top of Git that stores repositories online for team collaboration.
Identity Configuration:
- git --version : check installed Git version.
- git config --global user.name "Your Name"
- git config --global user.email "you@example.com"
- git config --list : verify configuration settings.
Authentication with GitHub:
- Option A — Personal Access Token (PAT): GitHub no longer accepts account passwords over HTTPS; you generate a PAT in Developer Settings and paste it when prompted for a password.
- Option B — SSH Key (Recommended):
  1. ssh-keygen -t ed25519 -C "you@example.com" generates two files: ~/.ssh/id_ed25519 (private key — NEVER share this) and ~/.ssh/id_ed25519.pub (public key — uploaded to GitHub Settings -> SSH and GPG keys).
  2. Test connection with: ssh -T git@github.com.`,
				},
				{
					ID:              "obj-4-2",
					TopicID:         "topic-4-git",
					Code:            "M1.W3.OBJ2",
					Title:           "Working Directory, Staging Area, Commits, Remotes & Core Git Commands",
					Statement:       "Use the three-step Git workflow and distinguish between clone, pull, fetch, push, and fork.",
					SourceReference: "Module 1, Week 3 — Git & Version Control (Slides 7–11, 13)",
					Ordering:        2,
					CurriculumExcerpt: `Git Areas & Terms:
- Repository (Repo): project folder tracked by Git containing a hidden .git/ directory with full history.
- Working Directory: your desk where you actively edit files.
- Staging Area (Index): the "Ready" tray where changes are prepared for the next snapshot via git add.
- Commit: an immutable snapshot of staged changes at a point in time with a unique hash ID and descriptive message.
- HEAD: pointer to the current commit or branch you are currently checked out on ("You are here").
- Remote (origin): link to a shared copy of the repository hosted online (GitHub/GitLab/Bitbucket).
Core Commands:
- git init : initialize a new local Git repository.
- git clone <url> : download a full local copy of a remote repository.
- git status : inspect modified, staged, and untracked files.
- git add <file> or git add . : move changes from working directory to staging area.
- git commit -m "message" : save staged snapshot to local commit history.
- git log : view commit history.
- git remote -v : list configured remote URLs; git remote add origin <url> links a local repo to a remote.
- git push origin main : upload local commits to the remote repository.
- git fetch vs. git pull : git fetch downloads remote changes without modifying your working files/branch, whereas git pull downloads remote changes AND merges them into your current branch.
- git fork (on GitHub/GitLab): creates an independent server-side copy of someone else's repository under your own account.`,
				},
				{
					ID:              "obj-4-3",
					TopicID:         "topic-4-git",
					Code:            "M1.W3.OBJ3",
					Title:           "Branching, Merging & Resolving Merge Conflicts",
					Statement:       "Create and switch branches for isolated feature development and resolve merge conflicts.",
					SourceReference: "Module 1, Week 3 — Git & Version Control (Slides 8, 10–13)",
					Ordering:        3,
					CurriculumExcerpt: `Branches allow developers to work on features or bug fixes in an isolated line of development without breaking the main codebase.
- git branch : list branches; git branch feature/login creates a new branch.
- git checkout feature/login (or git switch feature/login / git checkout -b feature/login): switch HEAD to that branch.
- git merge <branch> : combine changes from another branch into the current branch.
Merge Conflicts:
- Occur when two branches modify the exact same lines in the same file (or one deletes a file another modifies) and Git cannot automatically determine which change to keep.
- Resolution workflow: open the conflicted file, edit the conflict markers to choose or combine the desired code, stage the resolved file with git add, and complete the merge with git commit.`,
				},
			},
		},
		{
			ID:           "topic-5-compute-iam",
			Slug:         "core-compute-and-iam",
			ModuleNumber: 2,
			ModuleTitle:  "Module 2: Cloud Architecture, Services & AI",
			WeekCode:     "M2_W1",
			Title:        "5. Core Compute & IAM",
			Description:  "Virtual machines (anatomy, regions vs. zones, machine types, boot disks, lifecycle & billing states), compute spectrum (VMs, containers, serverless, managed platforms), pricing models (On-Demand, Reserved/Committed, Spot/Preemptible, Free Tier), cost alerts, and IAM (Principals, Users, Groups, Service Accounts, Authentication vs. Authorization, AWS JSON Policies vs. GCP Role Bindings, Least Privilege).",
			SourcePdf:    "M2_W1_Core Compute & IAM.pdf",
			Ordering:     5,
			Objectives: []models.LearningObjective{
				{
					ID:              "obj-5-1",
					TopicID:         "topic-5-compute-iam",
					Code:            "M2.W1.OBJ1",
					Title:           "Virtual Machine Anatomy, Regions vs. Zones & Compute Spectrum",
					Statement:       "Configure virtual machines, distinguish regions from availability zones, and compare VMs, containers, managed platforms, and serverless.",
					SourceReference: "Module 2, Week 1 — Core Compute & IAM (Slides 2–3, 8)",
					Ordering:        1,
					CurriculumExcerpt: `Anatomy of a Cloud Virtual Machine (VM):
- Region: geographic area (e.g., europe-west1, us-east-1). Determines latency, data residency compliance, and pricing.
- Zone (Availability Zone): an isolated data center within a region. Spreading VMs across multiple zones protects against data-center failures.
- Machine Type: CPU and RAM allocation (e.g., e2-micro, t3.micro; general-purpose, compute-optimized, memory-optimized).
- Image (AMI / OS Image): the operating system template (e.g., Ubuntu 22.04 LTS, Debian, Amazon Linux, Windows Server).
- Boot Disk: persistent block storage volume where the OS lives.
- Network & Firewall: VPC/subnet placement, public IP assignment, and firewall rules (e.g., Allow HTTP traffic on port 80, SSH on port 22).
Compute Service Spectrum (from most control/management to least):
1. Virtual Machines (AWS EC2 / GCP Compute Engine): full OS control; you manage OS patches and software.
2. Containers (AWS ECS, EKS / GCP Cloud Run, GKE): app + dependencies packaged together sharing host OS kernel; faster startup, less OS maintenance.
3. Managed Platforms (AWS Elastic Beanstalk / GCP App Engine): upload application code; platform handles servers and scaling.
4. Serverless Functions (AWS Lambda / GCP Cloud Functions): event-driven functions billed strictly per execution with zero server management.`,
				},
				{
					ID:              "obj-5-2",
					TopicID:         "topic-5-compute-iam",
					Code:            "M2.W1.OBJ2",
					Title:           "VM Lifecycle Billing, Pricing Models (On-Demand, Reserved, Spot) & Cost Traps",
					Statement:       "Explain what resources incur charges when a VM is stopped vs. deleted, select appropriate cloud pricing models, and configure cost alerts.",
					SourceReference: "Module 2, Week 1 — Core Compute & IAM (Slides 2, 10–11)",
					Ordering:        2,
					CurriculumExcerpt: `VM Lifecycle & Billing Impact:
- Running: you pay for CPU, RAM, AND attached disk storage (plus any paid public IP).
- Stopped: you stop paying for CPU and RAM, BUT you STILL PAY for the attached persistent boot disk (and any static/Elastic IP).
- Deleted / Terminated: VM and its auto-deleted boot disk are removed and billing stops.
Compute Pricing Models:
1. On-Demand: pay by the second/hour with no long-term commitment. Most flexible, highest hourly rate; best for short-term testing, learning, and unpredictable spiky traffic.
2. Reserved Instances / Committed Use Discounts: commit to 1 or 3 years of steady usage for a 30–70% discount; best for predictable always-on workloads like production databases.
3. Spot / Preemptible VMs: use provider's spare capacity at up to ~90% discount, but the provider can reclaim/terminate the instance with short notice. Best for fault-tolerant, interruptible batch jobs and data processing—NEVER for critical databases or stateful single-node servers.
Cost Alerts & Common Cost Traps:
- Billing budgets & alerts notify you via email at thresholds (e.g., 50%, 90%, 100%), but alerts DO NOT automatically stop or delete resources unless automated shutdown is separately configured. Billing data can also lag by hours.
- Common cost traps: forgotten running VMs, unattached persistent disks left after VM deletion, idle static/Elastic IPs, NAT Gateways, and oversized instances.`,
				},
				{
					ID:              "obj-5-3",
					TopicID:         "topic-5-compute-iam",
					Code:            "M2.W1.OBJ3",
					Title:           "IAM Principals, Service Accounts, Policies vs. Roles & Least Privilege",
					Statement:       "Apply the Principle of Least Privilege using IAM users, groups, service accounts, AWS JSON policies, and GCP role bindings.",
					SourceReference: "Module 2, Week 1 — Core Compute & IAM (Slides 4–6, 9)",
					Ordering:        3,
					CurriculumExcerpt: `IAM (Identity and Access Management) controls WHO (Principal) can do WHAT (Role/Permission/Action) on WHICH resource.
Authentication vs. Authorization:
- Authentication: proving who you are (e.g., password, MFA, SSH key).
- Authorization: determining what actions you are allowed to perform once authenticated (IAM permissions).
Principal Types:
- User: a single human operator with login credentials.
- Group: a collection of users; permissions granted to the group are inherited by all members.
- Service Account (or IAM Role for EC2/Lambda): a non-human identity assigned to an application or VM so it can securely access cloud resources (e.g., reading from a storage bucket) without hardcoding personal access keys in code or GitHub.
AWS JSON Policies vs. GCP Role Bindings:
- In AWS, an IAM Policy is a JSON document containing Effect ("Allow" or "Deny"), Action (e.g., "s3:GetObject"), and Resource (e.g., "arn:aws:s3:::my-bucket/*") attached to a user, group, or role.
- In GCP, you bind a pre-built Role (a bundle of permissions, such as roles/storage.objectViewer) to a Principal (user/group/service account) on a specific Resource (e.g., one bucket).
Principle of Least Privilege:
- Grant only the minimum permissions required to perform the intended task on the specific resource—nothing more (e.g., granting Storage Object Viewer on one bucket rather than Project Owner). If credentials leak to GitHub, bots can spin up hundreds of crypto-mining VMs within minutes ($50,000+ bill).`,
				},
			},
		},
		{
			ID:           "topic-6-networking",
			Slug:         "cloud-networking",
			ModuleNumber: 2,
			ModuleTitle:  "Module 2: Cloud Architecture, Services & AI",
			WeekCode:     "M2_W2",
			Title:        "6. Cloud Networking",
			Description:  "Virtual Private Cloud (VPC), subnetting, CIDR notation (/32, /24, /16, /8), Public vs. Private subnets, Route Tables, Internet Gateway (IGW), NAT Gateway, Elastic IP (EIP), ENI, Stateful Security Groups vs. Stateless Network ACLs (NACLs), Bastion Hosts & Session Manager, VPC Endpoints, ALB vs. NLB, Route 53, VPC Peering, Transit Gateway, VPN vs. Direct Connect, and VPC Flow Logs.",
			SourcePdf:    "M2_W2_Networking & Security.pdf",
			Ordering:     6,
			Objectives: []models.LearningObjective{
				{
					ID:              "obj-6-1",
					TopicID:         "topic-6-networking",
					Code:            "M2.W2.OBJ1",
					Title:           "VPC, Subnets, CIDR Notation & Public vs. Private Routing (IGW & NAT)",
					Statement:       "Design VPC subnet architectures using CIDR notation, route tables, Internet Gateways, and NAT Gateways.",
					SourceReference: "Module 2, Week 2 — Networking & Security (Slides 3–7, 12)",
					Ordering:        1,
					CurriculumExcerpt: `VPC (Virtual Private Cloud): A logically isolated private network within a public cloud provider's infrastructure. A VPC spans one Region across all Availability Zones (AZs) in that region, while each Subnet lives in a single AZ.
Subnetting & CIDR (Classless Inter-Domain Routing):
- Splits a VPC into smaller isolated sub-networks to prevent security lateral movement and broadcast traffic jams.
- Format: Base IP + Routing Prefix (e.g., 10.0.0.0/24). The prefix indicates how many of the 32 IPv4 bits are fixed for the network. Smaller prefix number = MORE IP addresses (each step down doubles the IP count):
  - /32 = 1 IP address (locks all 32 bits; targets a single specific host IP).
  - /24 = 256 IP addresses (standard cloud subnet size).
  - /16 = 65,536 IP addresses (standard VPC network block size).
  - /8 = 16,777,216 IP addresses.
Public vs. Private Subnets & Route Tables:
- Every subnet has a Route Table with a built-in local route (e.g., 10.0.0.0/16 -> local) for internal VPC communication.
- A subnet is PUBLIC strictly because its Route Table has a route sending internet-bound traffic (0.0.0.0/0) to an Internet Gateway (IGW). Resources in a public subnet also need a Public IP or Elastic IP to talk through the IGW.
- A subnet is PRIVATE because its Route Table has NO direct route to an Internet Gateway. Backend databases and internal services live here.
- NAT Gateway: Placed in a PUBLIC subnet (with an Elastic IP) so instances in a PRIVATE subnet can initiate outbound connections to the internet (e.g., running sudo apt update or calling external APIs) while blocking any inbound connections from the internet. The private route table routes 0.0.0.0/0 -> nat-gateway-id. Note: NAT Gateways charge hourly plus per GB processed—a major lab cost trap!
- Elastic IP (EIP): A static public IPv4 address that persists across instance stop/start cycles (charges apply, especially when idle/unattached).
- Elastic Network Interface (ENI): Virtual network card holding an instance's private IP, MAC address, and security groups.`,
				},
				{
					ID:              "obj-6-2",
					TopicID:         "topic-6-networking",
					Code:            "M2.W2.OBJ2",
					Title:           "Security Groups (Stateful) vs. Network ACLs (Stateless) & Bastion Hosts",
					Statement:       "Contrast instance-level Security Groups with subnet-level NACLs and configure secure Bastion Host / SG chaining rules.",
					SourceReference: "Module 2, Week 2 — Networking & Security (Slides 8–9, 11–13)",
					Ordering:        2,
					CurriculumExcerpt: `Security Group (SG) vs. Network Access Control List (NACL):
1. Security Group (Instance / ENI level virtual firewall):
   - Stateful: if an inbound request is allowed in, its outbound reply is automatically allowed back out regardless of outbound rules (and vice versa).
   - Allow rules ONLY: you cannot write explicit Deny rules.
   - Default state: denies all inbound traffic, allows all outbound traffic.
   - Evaluation: evaluates all rules together before deciding.
   - Supports Security Group Chaining: a rule's source can be another Security Group ID (e.g., Database SG allows port 3306 ONLY from the Web Server SG—the gold-standard least-privilege pattern).
2. Network ACL (NACL — Subnet level firewall):
   - Stateless: does not track connection state; you must explicitly allow both inbound traffic AND outbound ephemeral return ports (1024–65535).
   - Supports both Allow AND Deny rules (useful for explicitly blocking a malicious IP range).
   - Evaluation: rules are evaluated in numeric order from lowest rule number to highest; first match wins.
   - Default NACL allows all inbound and outbound traffic, whereas a newly created custom NACL denies everything by default.
Bastion Host (Jump Box) & Session Manager:
- A Bastion Host is a hardened VM in a Public Subnet used solely to SSH (port 22) or RDP (port 3389) into private-subnet servers.
- Outer Wall (Bastion SG): allow port 22 inbound ONLY from your specific admin IP (e.g., 203.0.113.50/32) — NEVER 0.0.0.0/0.
- Inner Wall (Private Server SG): allow port 22 inbound ONLY from the Bastion Host's Security Group ID or private IP.
- AWS Systems Manager (SSM) Session Manager: modern alternative providing browser/CLI shell access to private instances with NO open inbound SSH port 22 and no bastion VM required.`,
				},
				{
					ID:              "obj-6-3",
					TopicID:         "topic-6-networking",
					Code:            "M2.W2.OBJ3",
					Title:           "VPC Endpoints, Load Balancers (ALB/NLB), Route 53, Peering, Transit Gateway & Flow Logs",
					Statement:       "Select appropriate cloud networking services for service connectivity, load balancing, DNS, hybrid interconnect, and traffic auditing.",
					SourceReference: "Module 2, Week 2 — Networking & Security (Slides 10, 12)",
					Ordering:        3,
					CurriculumExcerpt: `Advanced Cloud Networking Components:
- VPC Endpoints: Allow private subnet instances to privately reach cloud services (like S3 or DynamoDB) without traversing the public internet or paying for a NAT Gateway. Gateway Endpoints (S3, DynamoDB) are free; Interface Endpoints (AWS PrivateLink) cost hourly.
- Elastic Load Balancer (ELB): Distributes incoming traffic across multiple healthy instances across AZs.
  - Application Load Balancer (ALB): Layer 7 (HTTP/HTTPS); routes based on URL path, hostname, or HTTP headers.
  - Network Load Balancer (NLB): Layer 4 (TCP/UDP); ultra-high performance and low latency.
- Route 53: Managed DNS service that translates domain names (app.example.com) into IP addresses or load balancer endpoints with health checks and routing policies.
- VPC Peering: Direct private IP connection between two VPCs. Non-transitive (if VPC A peers with B, and B peers with C, A CANNOT talk to C through B) and CIDR blocks must not overlap.
- Transit Gateway: Central hub-and-spoke router connecting many VPCs and on-premises networks, solving complex VPC peering meshes at scale.
- Site-to-Site VPN vs. Direct Connect: VPN creates an encrypted tunnel over the public internet; Direct Connect provides a dedicated physical fiber connection from on-premises to the cloud.
- VPC Flow Logs: Captures metadata of accepted and rejected network traffic entering/leaving network interfaces (ENIs)—the primary diagnostic tool for answering "why is my traffic being blocked by a NACL or Security Group?".`,
				},
			},
		},
		{
			ID:           "topic-7-storage",
			Slug:         "cloud-storage",
			ModuleNumber: 2,
			ModuleTitle:  "Module 2: Cloud Architecture, Services & AI",
			WeekCode:     "M2_W1_W2",
			Title:        "7. Cloud Storage",
			Description:  "Block storage (AWS EBS / GCP Persistent Disk) vs. Object storage (AWS S3 / GCP Cloud Storage), bucket naming & regional placement, private-by-default security, object immutability, Linux virtual storage & tmpfs, storage access patterns via IAM roles and VPC Gateway Endpoints, and storage cost traps.",
			SourcePdf:    "M2_W1_Core Compute & IAM.pdf & M2_W2_Networking & Security.pdf",
			Ordering:     7,
			Objectives: []models.LearningObjective{
				{
					ID:              "obj-7-1",
					TopicID:         "topic-7-storage",
					Code:            "M2.W1.OBJ4",
					Title:           "Block Storage (EBS / Persistent Disk) vs. Object Storage (S3 / Cloud Storage)",
					Statement:       "Compare block storage and object storage across attachment model, performance, mutability, cost, and use cases.",
					SourceReference: "Module 2, Week 1 — Core Compute & IAM (Slides 2, 7, 9, 11)",
					Ordering:        1,
					CurriculumExcerpt: `Block Storage vs. Object Storage:
1. Block Storage (AWS EBS / GCP Persistent Disk):
   - Raw disk volume attached directly to a single Virtual Machine (like a physical laptop hard drive).
   - Formatted with a filesystem (e.g., ext4, xfs) and mounted by the VM's OS.
   - Supports low-latency random read/write in-place updates; ideal for Operating Systems (boot disks) and transactional Databases.
   - Higher cost per GB; continues billing even when the attached VM is in the Stopped state, and unattached ("orphaned") disks continue billing until deleted.
2. Object Storage (AWS S3 / GCP Cloud Storage):
   - Stores files as whole objects inside flat containers called Buckets, accessed over HTTP/HTTPS via URL or REST API (like Google Drive).
   - Not mounted as a block boot drive; you cannot edit a single byte in-place—modifying a file replaces the entire object.
   - Practically unlimited capacity, low cost per GB, and high durability.
   - Ideal for backups, user uploads (images, videos, documents), application logs, static website assets, and ML datasets.`,
				},
				{
					ID:              "obj-7-2",
					TopicID:         "topic-7-storage",
					Code:            "M2.W1.OBJ5",
					Title:           "Bucket Security, Private-by-Default Access, IAM Storage Roles & VPC Endpoints",
					Statement:       "Secure object storage buckets using public-access prevention, least-privilege IAM policies/roles, and private VPC Gateway Endpoints.",
					SourceReference: "Module 2, Week 1 (Slides 5, 9) & Module 2, Week 2 (Slide 10)",
					Ordering:        2,
					CurriculumExcerpt: `Cloud Storage Bucket Security & Access Best Practices:
- Globally Unique Naming & Region Selection: Bucket names must be globally unique across all users; place buckets in the same region as compute workloads to minimize latency and cross-region data transfer costs.
- Private by Default: Always keep "Prevent public access" / "Block all public access" enabled unless explicitly hosting public static assets. Opening an object's URL in an unauthenticated browser window should return Access Denied.
- Least-Privilege Access: Grant fine-grained permissions such as GCP "Storage Object Viewer" on a single bucket or an AWS IAM JSON policy with Effect: "Allow", Action: "s3:GetObject", Resource: "arn:aws:s3:::my-bucket/*".
- Service Account / Execution Role Access: VMs and serverless functions should read/write buckets using an attached Service Account or IAM Role rather than long-lived user access keys.
- Private Network Access via VPC Gateway Endpoint: Private-subnet VMs can read/write S3 buckets without a NAT Gateway or public internet traversal by attaching a free S3 Gateway VPC Endpoint to the private route table.`,
				},
			},
		},
		{
			ID:           "topic-8-databases",
			Slug:         "databases",
			ModuleNumber: 2,
			ModuleTitle:  "Module 2: Cloud Architecture, Services & AI",
			WeekCode:     "M2_W3",
			Title:        "8. Databases",
			Description:  "SQL (relational: MySQL, PostgreSQL, Aurora) vs. NoSQL (non-relational: DynamoDB, MongoDB, Firestore), Self-Managed DB on VM vs. Managed DB (AWS RDS / GCP Cloud SQL), RDS security & Multi-AZ failover, DynamoDB partition/sort keys, Query vs. Scan vs. Filter, and database cost & troubleshooting.",
			SourcePdf:    "M2_W3_ Databases, Serverless & Application Services.pdf",
			Ordering:     8,
			Objectives: []models.LearningObjective{
				{
					ID:              "obj-8-1",
					TopicID:         "topic-8-databases",
					Code:            "M2.W3.OBJ1",
					Title:           "SQL (Relational) vs. NoSQL (Non-Relational) Databases",
					Statement:       "Compare relational SQL databases with NoSQL databases across schema rigidity, scaling, consistency, and query patterns.",
					SourceReference: "Module 2, Week 3 — Databases, Serverless & Application Services (Slides 2–3)",
					Ordering:        1,
					CurriculumExcerpt: `SQL (Relational Databases):
- Stores data in structured tables with rows, columns, and a strict predefined schema (like an Excel workbook with strict columns and linked sheets).
- Tables link through primary and foreign keys and are queried using SQL (Structured Query Language: CREATE TABLE, INSERT INTO, SELECT, WHERE, JOIN).
- Strong consistency and ACID transactions. Examples: MySQL, PostgreSQL, MariaDB, SQL Server, Oracle, Amazon Aurora.
NoSQL (Non-Relational Databases):
- Flexible structures (key-value, document, graph, wide-column) with no fixed table-wide schema—each item in the same table can have different attributes (like a filing cabinet where each folder holds different items).
- Scales horizontally and delivers ultra-fast, predictable key-based lookups. Examples: Amazon DynamoDB, MongoDB, Google Cloud Firestore.
- Neither is universally "better"; SQL excels at relational integrity and complex joins, while NoSQL excels at high-velocity key-value/document lookups and flexible schemas.`,
				},
				{
					ID:              "obj-8-2",
					TopicID:         "topic-8-databases",
					Code:            "M2.W3.OBJ2",
					Title:           "Self-Managed vs. Managed Databases (AWS RDS / Cloud SQL) & Private Security",
					Statement:       "Distinguish provider vs. customer responsibilities in managed databases and lock down RDS access using private subnets and Security Group chaining.",
					SourceReference: "Module 2, Week 3 — Databases, Serverless & Application Services (Slides 4–7)",
					Ordering:        2,
					CurriculumExcerpt: `Self-Managed vs. Managed Databases (AWS RDS / GCP Cloud SQL):
- Self-Managed (installing MySQL on an EC2 VM): You must manually handle OS patching, DB software upgrades, automated backups, point-in-time recovery, high-availability failover, and replication.
- Managed Database (RDS / Cloud SQL): The cloud provider automates backups, point-in-time recovery, OS & DB patching, Multi-AZ failover (maintaining a synchronous standby replica in another Availability Zone), read replicas, and monitoring.
- What the customer STILL owns in RDS: schema & query design, Security Group rules, database credentials (master username/password), and cost management.
RDS Security & Operational Rules:
- Place RDS in a PRIVATE subnet with "Public access: No".
- Lock down the RDS Security Group (e.g., rds-lab-sg) to allow inbound port 3306 (MySQL/Aurora) or 5432 (PostgreSQL) ONLY from the Application VM's Security Group ID—never from 0.0.0.0/0.
- Connect from the authorized app VM using: mysql -h <rds-endpoint> -u admin -p.
- Cost Warning: RDS bills continuously while running even when idle. If you stop an RDS instance, AWS automatically restarts it after 7 days! Disable storage autoscaling in labs and delete final snapshots when cleaning up test databases.`,
				},
				{
					ID:              "obj-8-3",
					TopicID:         "topic-8-databases",
					Code:            "M2.W3.OBJ3",
					Title:           "DynamoDB Data Modeling, Partition Keys & Query vs. Scan",
					Statement:       "Model items in Amazon DynamoDB and contrast efficient Partition-Key Query operations with full-table Scan operations.",
					SourceReference: "Module 2, Week 3 — Databases, Serverless & Application Services (Slides 8–10)",
					Ordering:        3,
					CurriculumExcerpt: `Amazon DynamoDB (Managed Serverless NoSQL Database):
- Primary Key: Every table requires a Partition Key (e.g., id of type String) and optionally a Sort Key. The data type of the Partition Key must match exactly (e.g., passing String "1" vs. Number 1 causes a type mismatch validation error).
- Capacity Modes: On-Demand (pay per read/write request with automatic scaling) or Provisioned capacity.
Reading Data in DynamoDB — Query vs. Scan:
1. Query (Fast & Cost-Efficient): Looks up items directly using the Partition Key (e.g., id = "2"). Reads only the targeted partition; always prefer Query in production.
2. Scan (Slow & Expensive): Reads EVERY single item in the entire table. Even if you attach a Scan Filter (e.g., course = "DevOps"), DynamoDB still scans the whole table first and filters the results afterward.
CLI Commands (Typed JSON format):
- aws dynamodb put-item --table-name students --item '{"id":{"S":"4"},"name":{"S":"Chidi"},"course":{"S":"Security"}}'
- aws dynamodb get-item --table-name students --key '{"id":{"S":"4"}}'
- aws dynamodb scan --table-name students`,
				},
			},
		},
		{
			ID:           "topic-9-serverless",
			Slug:         "serverless-and-application-services",
			ModuleNumber: 2,
			ModuleTitle:  "Module 2: Cloud Architecture, Services & AI",
			WeekCode:     "M2_W3",
			Title:        "9. Serverless & Application Services",
			Description:  "Serverless computing principles, AWS Lambda / GCP Cloud Functions (handler, event, context, triggers, execution roles, 15-minute timeout, billed duration, cold starts, CloudWatch logs), and managed Application Services (App Hosting: Elastic Beanstalk, App Runner, Fargate, Cloud Run; API Gateway; Messaging: SQS, SNS, EventBridge, Pub/Sub; Caching: ElastiCache, Memorystore; User Auth: Cognito, Firebase Auth; CDN: CloudFront; Workflows: Step Functions).",
			SourcePdf:    "M2_W3_ Databases, Serverless & Application Services.pdf",
			Ordering:     9,
			Objectives: []models.LearningObjective{
				{
					ID:              "obj-9-1",
					TopicID:         "topic-9-serverless",
					Code:            "M2.W3.OBJ4",
					Title:           "Serverless Computing & AWS Lambda Architecture, Limits & Billing",
					Statement:       "Explain serverless execution, AWS Lambda triggers, execution roles, timeouts, cold starts, and CloudWatch log reports.",
					SourceReference: "Module 2, Week 3 — Databases, Serverless & Application Services (Slides 11–13, 16)",
					Ordering:        1,
					CurriculumExcerpt: `Serverless & AWS Lambda (or GCP Cloud Functions):
- Serverless means physical servers still exist, but the cloud provider handles all provisioning, patching, idle capacity, and scaling. You upload function code and pay strictly when it executes ($0 when idle).
Core Lambda Components:
- Function Handler: e.g., Python def lambda_handler(event, context): where event contains input payload data (use event.get("name", "World") to avoid KeyError on missing keys) and context holds runtime metadata. Remember to click Deploy after editing code in the console!
- Trigger: the event source that invokes the function (e.g., API Gateway HTTP request, S3 object upload, scheduled EventBridge/cron timer, SQS queue message).
- Execution Role: the IAM role attached to the Lambda function granting it least-privilege permissions to access other services (e.g., writing logs to CloudWatch Logs or reading a DynamoDB table).
- Limits & Characteristics: maximum execution timeout is 15 minutes (900 seconds) per invocation; configurable memory allocation (which also scales CPU).
- Cold Start: the initial latency delay when a serverless function is invoked for the first time (or after inactivity) while the provider provisions a container environment and initializes the runtime.
- Monitoring & Billing: CloudWatch Logs records START, END, and REPORT lines for every invocation, showing Duration and Billed Duration (billed per request + GB-seconds of compute duration).`,
				},
				{
					ID:              "obj-9-2",
					TopicID:         "topic-9-serverless",
					Code:            "M2.W3.OBJ5",
					Title:           "Managed Application Services (Queues, Pub/Sub, API Gateway, Caching, Auth, CDN, Workflows)",
					Statement:       "Select managed cloud application services for decoupling, API management, caching, end-user authentication, content delivery, and orchestration.",
					SourceReference: "Module 2, Week 3 — Databases, Serverless & Application Services (Slides 14–15)",
					Ordering:        2,
					CurriculumExcerpt: `Managed Application Service Categories:
1. App Hosting (run code/containers without managing VMs):
   - AWS Elastic Beanstalk / GCP App Engine (upload application code, platform manages environment — medium control).
   - AWS ECS Fargate / AWS App Runner / GCP Cloud Run (run containers serverlessly — medium-low control).
   - AWS Amplify (host frontend web apps).
2. API Management (expose HTTP endpoints, handle authentication, rate throttling, and routing to Lambda/backends):
   - AWS API Gateway / GCP API Gateway & Apigee.
3. Messaging, Queues & Events (decouple microservices so if one component fails, messages wait safely in a queue instead of being lost):
   - AWS SQS (Simple Queue Service — point-to-point message queue).
   - AWS SNS (Simple Notification Service — pub/sub fan-out notifications).
   - AWS EventBridge (event bus routing) / GCP Pub/Sub.
4. In-Memory Caching (store frequently read data in RAM so the database is not hit on every request):
   - AWS ElastiCache (Redis/Memcached) / GCP Memorystore.
5. End-User Authentication & Identity (handles sign-up, login, passwords, and OAuth for your application's customers — distinct from IAM, which is for cloud engineers and infrastructure):
   - AWS Cognito / GCP Identity Platform & Firebase Auth.
6. Content Delivery Network — CDN (caches and serves static assets like images, video, and web pages from edge locations close to users for low latency):
   - AWS CloudFront / GCP Cloud CDN.
7. Workflow Orchestration (chains multiple serverless functions and steps into a stateful workflow with retries and branching):
   - AWS Step Functions / GCP Workflows.`,
				},
			},
		},
		{
			ID:           "topic-10-ai-mlops",
			Slug:         "cloud-ai-and-mlops",
			ModuleNumber: 2,
			ModuleTitle:  "Module 2: Cloud Architecture, Services & AI",
			WeekCode:     "M2_W4",
			Title:        "10. Cloud AI & MLOps",
			Description:  "Traditional AI vs. Generative AI (GenAI) & LLMs, Assist vs. Embed patterns, core LLM parameters (Prompt, System Prompt, Token, Context Window, Temperature, Hallucination), the 3 layers of Cloud AI services (AWS Bedrock, SageMaker, Rekognition/Transcribe/Comprehend; GCP Vertex AI; Azure OpenAI), RCTFC Prompt Engineering framework, MLOps lifecycle & Model Drift, and Responsible AI practices.",
			SourcePdf:    "M2_W4_ Cloud AI & MLOps Fundamentals.pdf",
			Ordering:     10,
			Objectives: []models.LearningObjective{
				{
					ID:              "obj-10-1",
					TopicID:         "topic-10-ai-mlops",
					Code:            "M2.W4.OBJ1",
					Title:           "Traditional AI vs. Generative AI, LLM Mechanics & Cloud Operations Use Cases",
					Statement:       "Distinguish Traditional AI from Generative AI, explain Assist vs. Embed patterns, and define tokens, context windows, temperature, system prompts, and hallucinations.",
					SourceReference: "Module 2, Week 4 — Cloud AI & MLOps Fundamentals (Slides 3–6)",
					Ordering:        1,
					CurriculumExcerpt: `Traditional AI vs. Generative AI (GenAI):
- Traditional AI: Analyzes existing data to predict a classification label or numeric value (e.g., "Is this email spam?").
- Generative AI (GenAI): Creates new content (text, code, images, audio, video) based on patterns learned from massive training datasets (e.g., "Write an incident summary or Bash script"). Large Language Models (LLMs like Claude, GPT, Gemini) predict the next token sequence.
Where AI Helps Cloud Engineers (Operations: log summaries, incident timelines, alert triage; Automation: drafting Bash scripts, Terraform IaC, runbooks; Support: ticket classification; Security & Cost: reviewing IAM policies, spotting anomalies):
- Assist Mode: A human engineer uses an interactive chat or console assistant to speed up their daily tasks.
- Embed Mode: An application, CI/CD pipeline, or Lambda function programmatically calls an AI API (e.g., automatically summarizing failed deployment logs or classifying incoming support tickets into JSON).
Core LLM Concepts:
- Prompt: The user input question or instructions sent to the model.
- System Prompt: Hidden standing instructions set by the developer that shape the model's persona, rules, and output format across every reply.
- Token: A chunk of text (~3/4 of a word; "Cloud computing" is 2–3 tokens). Cloud AI APIs meter and bill per input token AND output token.
- Context Window: The model's short-term memory limit—the maximum combined tokens (prompt + conversation history + reply) it can process at once.
- Temperature: The randomness control dial. Low temperature (near 0) produces deterministic, consistent outputs (best for generating code, JSON, and cloud configs); high temperature produces creative, varied outputs.
- Hallucination: Because LLMs optimize for what sounds plausible rather than verified truth, they can confidently invent non-existent CLI flags, fake API fields, or cloud services. Always verify generated commands against official docs or --help before execution.`,
				},
				{
					ID:              "obj-10-2",
					TopicID:         "topic-10-ai-mlops",
					Code:            "M2.W4.OBJ2",
					Title:           "Three Layers of Cloud AI Services & The RCTFC Prompt Framework",
					Statement:       "Map AWS, GCP, and Azure AI services across the three cloud AI layers and write structured cloud operations prompts using the RCTFC framework.",
					SourceReference: "Module 2, Week 4 — Cloud AI & MLOps Fundamentals (Slides 7–10, 14)",
					Ordering:        2,
					CurriculumExcerpt: `Three Layers of Cloud AI Services:
1. Infrastructure Layer (Raw compute — GPUs, TPUs, custom ML chips for training your own models from scratch; "buying raw ingredients").
2. Platform Layer (Managed environments to build, train, evaluate, deploy, and monitor custom ML models; "a fully equipped kitchen"):
   - AWS: Amazon SageMaker
   - GCP: Vertex AI
   - Azure: Azure Machine Learning
3. AI APIs & Foundation Models Layer (Ready-made models called via API with little/no ML training required; "ordering takeaway"):
   - Foundation Models via API: AWS Bedrock, GCP Vertex AI (Model Garden / Gemini), Azure OpenAI Service / Azure AI Foundry.
   - Prebuilt Task APIs (Vision, Speech, Text):
     - AWS: Amazon Rekognition (vision), Amazon Transcribe (speech-to-text), Amazon Comprehend (natural language/text).
     - GCP: Vision AI, Speech-to-Text, Natural Language API.
     - Azure: Azure AI Vision, Azure AI Speech, Azure AI Language.
The RCTFC Prompt Engineering Framework for Cloud Operations:
- R — Role: Define the persona (e.g., "You are a senior Linux and AWS cloud support engineer").
- C — Context: Supply the exact environment details, error logs, OS version, and architecture (e.g., "Ubuntu 24.04 VM on AWS behind an ALB returning 502 Bad Gateway; here are the last 20 log lines").
- T — Task: Specify the exact action required (e.g., "Summarize the log in 3 bullets, state the root cause, and list 3 diagnostic commands").
- F — Format: Define the output structure (e.g., "Return valid JSON only with keys category, priority, summary" or "a Markdown table with columns Step, Command, Expected Output").
- C — Constraints: Set guardrails (e.g., "Do not invent CLI flags; ask if context is insufficient; keep under 150 words").`,
				},
				{
					ID:              "obj-10-3",
					TopicID:         "topic-10-ai-mlops",
					Code:            "M2.W4.OBJ3",
					Title:           "MLOps Lifecycle, Model Drift & Responsible AI in Production",
					Statement:       "Explain the MLOps lifecycle loop, contrast MLOps with DevOps, and enforce the five Responsible AI practices and least privilege.",
					SourceReference: "Module 2, Week 4 — Cloud AI & MLOps Fundamentals (Slides 11–15)",
					Ordering:        3,
					CurriculumExcerpt: `MLOps (DevOps for Machine Learning):
- The MLOps Lifecycle Loop: Data -> Prepare -> Train -> Evaluate -> Register -> Deploy -> Monitor -> Retrain.
- How MLOps differs from traditional DevOps:
  1. You must version DATA and MODELS (in a Model Registry), not just source code.
  2. Models experience Model Drift: even if code never changes, real-world data shifts over time, causing model accuracy to degrade silently until monitored and retrained.
  3. Training must be reproducible (same dataset + hyperparameters = consistent model).
  4. Testing evaluates statistical model quality and bias, not just "does the binary compile and run".
- Key MLOps Terms:
  - Feature: an input variable/attribute the model learns from.
  - Training vs. Inference: Training is learning patterns from historical data; Inference is using a trained model in production to make predictions on new data.
  - Model Registry: a versioned catalog of trained models.
  - Canary Release: routing a small percentage (e.g., 5%) of live traffic to a new model version first to verify performance before full rollout.
Five Responsible AI Practices for Cloud Engineers:
1. Protect Privacy: Never paste passwords, IAM access keys, customer PII, or private internal URLs into AI prompts; sanitize with placeholders like <API_KEY> or customer-123, and prefer enterprise cloud AI services (Bedrock, Vertex AI) that do not train public models on your data.
2. Check for Accuracy: Verify every AI-suggested command (especially destructive ones like aws s3 rm s3://company-logs --recursive) against --help or official documentation and test in a sandbox first.
3. Keep a Human in the Loop: Let AI draft, but require human review and approval for high-impact/irreversible actions (deleting resources, modifying IAM policies, opening Security Groups, spending budget).
4. Watch for Bias and Unfair Outcomes: Test classification/prioritization models across diverse cases to ensure fair treatment.
5. Be Accountable and Transparent + Least Privilege for AI Agents: Log prompts and AI actions for auditability, and grant AI functions/agents strictly the minimum IAM permissions needed (a log-summarizer Lambda must NEVER have s3:DeleteObject or bucket deletion permissions).`,
				},
			},
		},
	}
}

