
pipeline {
    agent any

    stages {
        stage('Lint and Test') {
            steps {
                sh '''
                    docker run --rm \
                        -v "$PWD:/app" \
                        -w /app \
                        golang:1.27.2-alpine \
                        sh -c '
                            set -e
                            test -z "$(gofmt -l .)"
                            go vet ./...
                            go test -race ./...
                        '
                '''
            }
        }

        stage('Build') {
            steps {
                sh '''
                    docker build -t threat-service:${BUILD_NUMBER} .
                '''
            }
        }
    }
}

