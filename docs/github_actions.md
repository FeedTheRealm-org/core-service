# GitHub Actions (CI/CD)

The repository uses GitHub Actions for Continuous Integration and Continuous Deployment. The workflow definitions are located in `.github/workflows/`.

## Workflows Overview

- **`precommit-check.yml`**: Runs linting, formatting checks, and basic unit validations on pull requests to ensure code quality before merging.
- **`ci-cd.yml`**: A combined pipeline that handles building pushing to ECR and deploying.
- **`build.yml`**: Compiles the application and generates the Docker images to guarantee the build succeeds on the target branches.
- **`deploy.yml`**: Manages the deployment of the production-ready Docker image to the corresponding infrastructure via SSM command.

## Actions Needed

If configuring a new fork or repository, ensure you set up the required repository secrets in GitHub (such as registry strings, deployment keys, and instances targetting).
