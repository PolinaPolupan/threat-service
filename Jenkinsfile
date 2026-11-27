
pipeline {
    agent any

    environment {
        IMAGE_NAME = 'ghcr.io/polinapolupan/threat-service'
DEPLOY_STARTED = 'false'
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


stage('Deploy') {
    steps {
        withCredentials([sshUserPrivateKey(
            credentialsId: 'deploy-ssh-key',
            keyFileVariable: 'SSH_KEY',
            usernameVariable: 'SSH_USER'
        )]) {
            sh '''
                ssh -i "$SSH_KEY" \
                    "$SSH_USER@192.168.1.5" \
                    "cd ~/threat-service-deploy && \
                     grep '^IMAGE_TAG=' .env > .previous_image"
            '''

            script {
                env.DEPLOY_STARTED = 'true'
            }

            sh '''
                ssh -i "$SSH_KEY" \
                    "$SSH_USER@192.168.1.5" \
                    "cd ~/threat-service-deploy && \
                     sed -i 's/^IMAGE_TAG=.*/IMAGE_TAG=${GIT_COMMIT}/' .env && \
                     docker compose --env-file .env -f docker-compose.deploy.yml pull app && \
                     docker compose --env-file .env -f docker-compose.deploy.yml up -d"
            '''
        }
    }
}
stage('Health Check') {
    steps {
        withCredentials([sshUserPrivateKey(
            credentialsId: 'deploy-ssh-key',
            keyFileVariable: 'SSH_KEY',
            usernameVariable: 'SSH_USER'
        )]) {
            sh 'exit 1'
        }
    }
}

    }
post {
    failure {
        script {
            if (env.DEPLOY_STARTED == 'true') {
                echo 'Deployment failed. Rolling back...'

                withCredentials([sshUserPrivateKey(
                    credentialsId: 'deploy-ssh-key',
                    keyFileVariable: 'SSH_KEY',
                    usernameVariable: 'SSH_USER'
                )]) {
		sh '''
    ssh -i "$SSH_KEY" \
        "$SSH_USER@192.168.1.5" \
'set -e
         cd ~/threat-service-deploy
         test -s .previous_image
         grep -v "^IMAGE_TAG=" .env > .env.tmp
         cat .previous_image >> .env.tmp
         mv .env.tmp .env
         docker compose --env-file .env -f docker-compose.deploy.yml pull app
         docker compose --env-file .env -f docker-compose.deploy.yml up -d
         i=0
         while [ "$i" -lt 30 ]; do
             if curl -fsS http://localhost:8080/health; then
                 echo "Rollback health check passed"
                 exit 0
             fi
             i=$((i + 1))
             sleep 2
         done
         echo "Rollback health check failed"
         exit 1'
'''
                }
            } else {
                echo 'Deployment did not start. Rollback is not required.'
            }
        }
    }
}
}

