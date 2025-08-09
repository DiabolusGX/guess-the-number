module.exports = {
    apps: [
        {
            name: "guess-the-number-bot",
            script: "./main",
            cwd: "/app",
            instances: 1,
            exec_mode: "fork",
            autorestart: true,
            watch: false,
            max_memory_restart: "4096M",
            env: {
                NODE_ENV: "production",
                GTN_DEPLOYMENT_MODE: "production",
            },
            env_development: {
                NODE_ENV: "development",
                GTN_DEPLOYMENT_MODE: "local",
                GTN_LOGGING_LEVEL: "debug",
            },
            // Logging configuration
            log_file: "~/logs/gtn-bot/combined.log",
            out_file: "~/logs/gtn-bot/out.log",
            error_file: "~/logs/gtn-bot/pm2-error.log",
            log_date_format: "YYYY-MM-DD HH:mm:ss Z",

            // Process management
            min_uptime: "10s",
            max_restarts: 10,
            restart_delay: 4000,

            // Health monitoring
            health_check_grace_period: 3000,
            health_check_fatal_exceptions: true,

            // Advanced options
            kill_timeout: 5000,
            listen_timeout: 3000,

            // Environment-specific configurations
            merge_logs: true,
            combine_logs: true,

            // Graceful shutdown
            shutdown_with_message: true,
            wait_ready: true,

            // Clustering options (for future sharding expansion)
            increment_var: "PORT",

            // Source map support (if needed)
            source_map_support: false,

            // Process title
            proc_title: "gtn-bot",

            // Custom restart conditions
            ignore_watch: ["node_modules", "logs", "*.log"],

            // Graceful reload
            reload_signal: "SIGUSR2",
        },
    ],

    // Deployment configuration (optional)
    deploy: {
        production: {
            user: "deploy",
            host: ["production-server"],
            ref: "origin/main",
            repo: "git@github.com:diabolusgx/guess-the-number.git",
            path: "/var/www/guess-the-number",
            "post-deploy": "make prod-build",
            "pre-setup": "apt update && apt install docker.io docker-compose -y",
        },
        staging: {
            user: "deploy",
            host: ["staging-server"],
            ref: "origin/develop",
            repo: "git@github.com:diabolusgx/guess-the-number.git",
            path: "/var/www/guess-the-number-staging",
            "post-deploy": "make prod-build",
            env: {
                NODE_ENV: "staging",
            },
        },
    },
};
