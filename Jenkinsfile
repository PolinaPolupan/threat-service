
pipeline {
    agent any

    stages {   
stage('Lint and Test') {
    steps {
        sh '''
            docker run --rm \
                -v jenkins_jenkins_home:/jenkins_home \
                -w /jenkins_home/workspace/threat-service-pipeline \
                golang:1.27.2 \
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

