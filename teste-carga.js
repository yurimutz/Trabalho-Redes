import { check, sleep } from 'k6';
import http from 'k6/http';

// Configuração da "Força" do Teste
export const options = {
    vus: 1024,          // 1024 utilizadores simultâneos! (Carga Altíssima)
    // Como cada iteração demora (2s + 10 * 4s = 42 segundos), 
    // aumentamos a duração do teste para dar tempo de todos terminarem o ciclo
    duration: '1m30s',
};

const HOSTNAME = __ENV.HOSTNAME || 'localhost';

export default function () {
    // PASSO 1: O utilizador abre o site e pede o Catálogo
    let resCatalogo = http.get('http://192.168.100.111:8050/catalogo');

    check(resCatalogo, {
        'Catalogo carregou com sucesso': (r) => r.status === 200,
    });

    // O utilizador demora 2 segundos a olhar para o catálogo
    sleep(2);t

    // PASSO 2: O utilizador dá Play num vídeo!
    let resManifesto = http.get('http://192.168.100.111:8080/videos/Santi Cazorla, quando futebol vira arte./manifesto.mpd');

    check(resManifesto, {
        'Manifesto baixado': (r) => r.status === 200,
    });

    // PASSO 3: O reprodutor (DASH) pede 10 pedaços de vídeo
    for (let i = 1; i <= 10; i++) {
        // Monta o nome do chunk dinamicamente (ex: slice_0_1.m4s, slice_0_2.m4s)
        let pedaco = `chunk_1_${i}.m4s`;

        let resChunk = http.get(`http://192.168.100.111:8080/videos/Santi Cazorla, quando futebol vira arte./${pedaco}`);

        check(resChunk, {
            'Pedaco baixado': (r) => r.status === 200 || r.status === 206,
        });

        // O player espera 4 segundos exatos antes de pedir o próximo pedaço
    }
    for (let i = 1; i <= 5; i++) {
        sleep(4);
        let pedaco = `chunk_1_${i}.m4s`;

        let resChunk = http.get(`http://192.168.100.111:8080/videos/Santi Cazorla, quando futebol vira arte./${pedaco}`);

        check(resChunk, {
            'Pedaco baixado': (r) => r.status === 200 || r.status === 206,
        });
    }
}