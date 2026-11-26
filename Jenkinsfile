pipeline {
    agent any

    environment {
        IMAGE_NAME = 'ghcr.io/polinapolupan/threat-service'
    }

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
                    docker build \
                        -t ${IMAGE_NAME}:${GIT_COMMIT} \
                        .
                '''
            }
        }

        stage('Publish') {
            steps {
                withCredentials([usernamePassword(
                    credentialsId: 'ghcr-credentials',
                    usernameVariable: 'GHCR_USER',
                    passwordVariable: 'GHCR_TOKEN'
                )]) {
                    sh '''
                        set +x
                        echo "$GHCR_TOKEN" | docker login ghcr.io \
                            -u "$GHCR_USER" --password-stdin

                        docker push ${IMAGE_NAME}:${GIT_COMMIT}

                        docker logout ghcr.io
                    '''
                }
            }
        }
    }
}
