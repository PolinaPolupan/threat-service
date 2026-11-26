
pipeline {
    agent any

    stages {
        stage('Lint and Test') {
            steps {
                sh '''
                    set -e
                    test -z "$(gofmt -l .)"
                    go vet ./...
                    go test -race ./...
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

